package digitalocean

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

var prodTestRegions = []string{"tor1", "sfo3", "nyc3"}

func TestProdSSHKeyLagAgainstAPI(t *testing.T) {
	if os.Getenv("RUN_PROD_API_TEST") != "1" {
		t.Skip("set RUN_PROD_API_TEST=1 to hit api.digitalocean.com")
	}

	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		token = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}
	if token == "" {
		t.Fatal("DIGITALOCEAN_TOKEN or DIGITALOCEAN_ACCESS_TOKEN must be set")
	}

	client, err := godo.New(oauth2.NewClient(context.Background(), &APITokenSource{AccessToken: token}))
	require.NoError(t, err)

	ctx := context.Background()
	attempts := 5
	if v := os.Getenv("PROD_API_ATTEMPTS"); v != "" {
		fmt.Sscanf(v, "%d", &attempts)
	}
	parallel := 1
	if v := os.Getenv("PROD_API_PARALLEL"); v != "" {
		fmt.Sscanf(v, "%d", &parallel)
	}
	if parallel < 1 {
		parallel = 1
	}
	stopOn422 := os.Getenv("PROD_API_STOP_ON_422") != "0"

	var reproduced, recovered, firstTryOK int32
	var stop atomic.Bool
	var wg sync.WaitGroup
	errCh := make(chan error, parallel)

	perWorker := attempts
	if parallel > 1 {
		perWorker = (attempts + parallel - 1) / parallel
	}

	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 1; i <= perWorker; i++ {
				if stop.Load() {
					return
				}
				n := worker*1000 + i
				region := prodTestRegions[(worker+i)%len(prodTestRegions)]
				log.Printf("worker %d attempt %d region=%s: create key then immediately create droplet", worker, i, region)
				saw422, createdOnRetry, runErr := runProdSSHKeyLagAttempt(ctx, client, n, region)
				if runErr != nil {
					errCh <- fmt.Errorf("worker %d attempt %d region %s: %w", worker, i, region, runErr)
					stop.Store(true)
					return
				}
				if saw422 {
					atomic.AddInt32(&reproduced, 1)
					if createdOnRetry {
						atomic.AddInt32(&recovered, 1)
					}
					log.Printf("hit unhappy path worker=%d attempt=%d region=%s", worker, i, region)
					if stopOn422 {
						stop.Store(true)
						return
					}
					continue
				}
				atomic.AddInt32(&firstTryOK, 1)
			}
		}(w)
	}
	wg.Wait()
	close(errCh)
	for runErr := range errCh {
		require.NoError(t, runErr)
	}

	t.Logf("prod results: first_try_ok=%d reproduced_422=%d recovered_with_backoff=%d",
		firstTryOK, reproduced, recovered)
	if reproduced == 0 {
		t.Fatalf("race did not fire after %d scheduled attempts (%d parallel workers)", attempts, parallel)
	}
}

func runProdSSHKeyLagAttempt(ctx context.Context, client *godo.Client, n int, region string) (saw422 bool, createdOnRetry bool, err error) {
	pub, err := generateSSHPublicKey()
	if err != nil {
		return false, false, err
	}

	keyName := fmt.Sprintf("packer-esc26331-%d-%d", time.Now().UnixNano(), n)
	key, _, err := client.Keys.Create(ctx, &godo.KeyCreateRequest{
		Name:      keyName,
		PublicKey: pub,
	})
	if err != nil {
		return false, false, fmt.Errorf("create ssh key: %w", err)
	}
	log.Printf("created ssh key id=%d name=%s", key.ID, keyName)

	defer func() {
		delErr := retryWithBackoff(
			context.Background(),
			2*time.Minute,
			defaultRetryInitial,
			defaultRetryMaxBackoff,
			isNotFoundError,
			func() error {
				_, e := client.Keys.DeleteByID(context.Background(), key.ID)
				return e
			},
		)
		if delErr != nil {
			log.Printf("cleanup ssh key %d failed (delete manually): %v", key.ID, delErr)
		} else {
			log.Printf("deleted ssh key id=%d", key.ID)
		}
	}()

	req := &godo.DropletCreateRequest{
		Name:   fmt.Sprintf("packer-esc26331-%d-%d", time.Now().UnixNano(), n),
		Region: region,
		Size:   "s-1vcpu-1gb",
		Image:  godo.DropletCreateImage{Slug: "ubuntu-22-04-x64"},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: key.ID},
		},
		Tags: []string{"packer-esc26331-test"},
	}

	droplet, _, firstErr := client.Droplets.Create(ctx, req)
	if isInvalidSSHKeyError(firstErr) {
		saw422 = true
		log.Printf("reproduced 422 on first droplet create: %v", firstErr)

		retried := false
		err = retryWithBackoff(
			ctx,
			2*time.Minute,
			defaultRetryInitial,
			defaultRetryMaxBackoff,
			isInvalidSSHKeyError,
			func() error {
				var createErr error
				droplet, _, createErr = client.Droplets.Create(ctx, req)
				if isInvalidSSHKeyError(createErr) {
					retried = true
					log.Printf("retrying droplet create after invalid key identifiers: %v", createErr)
				}
				return createErr
			},
		)
		if err != nil {
			return saw422, retried, fmt.Errorf("droplet create still failing after backoff: %w", err)
		}
		createdOnRetry = true
		log.Printf("droplet create succeeded after exponential backoff")
	} else if firstErr != nil {
		return false, false, fmt.Errorf("droplet create: %w", firstErr)
	} else {
		log.Printf("droplet create succeeded on first try id=%d region=%s", droplet.ID, region)
	}

	if droplet == nil {
		return saw422, createdOnRetry, fmt.Errorf("droplet create returned empty droplet")
	}

	defer func() {
		if _, delErr := client.Droplets.Delete(context.Background(), droplet.ID); delErr != nil {
			log.Printf("cleanup droplet %d failed (destroy manually): %v", droplet.ID, delErr)
		} else {
			log.Printf("destroyed droplet id=%d", droplet.ID)
		}
	}()

	if !saw422 && os.Getenv("PROD_API_WAIT") != "1" {
		return saw422, createdOnRetry, nil
	}

	if err := waitForDropletState("active", droplet.ID, client, 6*time.Minute); err != nil {
		return saw422, createdOnRetry, fmt.Errorf("wait for active: %w", err)
	}

	err = retryWithBackoff(
		ctx,
		2*time.Minute,
		defaultRetryInitial,
		defaultRetryMaxBackoff,
		func(err error) bool {
			return strings.Contains(err.Error(), "IPv4") || isNotFoundError(err)
		},
		func() error {
			d, _, getErr := client.Droplets.Get(ctx, droplet.ID)
			if getErr != nil {
				return getErr
			}
			ip, ipErr := dropletIPv4(d, false)
			if ipErr != nil {
				log.Printf("waiting for ipv4: %v", ipErr)
				return ipErr
			}
			log.Printf("droplet ipv4=%s", ip)
			return nil
		},
	)
	if err != nil {
		return saw422, createdOnRetry, fmt.Errorf("wait for ipv4: %w", err)
	}

	return saw422, createdOnRetry, nil
}

func generateSSHPublicKey() (string, error) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub))), nil
}
