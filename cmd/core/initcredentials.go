package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"

	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
)

// initCredentialsInCluster is init-credentials as the Job runs it: with the
// pod's ServiceAccount, in the openfief namespace, and the system's random
// source.
func initCredentialsInCluster(ctx context.Context, log *slog.Logger) error {
	cl, err := inCluster()
	if err != nil {
		return err
	}
	return initCredentials(ctx, identitykube.NewSecretCreator(cl.client, cl.namespace), rand.Reader, log)
}

// initCredentials makes the platform's credentials through creator, with
// the bytes of random (D-216), and logs, for each Secret, its name and
// whether it was created or kept: never a value.
func initCredentials(ctx context.Context, creator identityapp.SecretCreator, random io.Reader, log *slog.Logger) error {
	outcomes, err := identityapp.InitPlatformCredentials(ctx, creator, random)
	for _, o := range outcomes {
		result := "kept"
		if o.Created {
			result = "created"
		}
		log.Info("platform credential", "secret", o.Name, "result", result)
	}
	if err != nil {
		return fmt.Errorf("platform credentials: %w", err)
	}
	return nil
}
