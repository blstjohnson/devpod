package workspace

import (
	"context"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/skevetter/devpod/e2e/framework"
)

var _ = framework.DevPodDescribe("devpod workspace rebind", func() {
	ginkgo.Context("rebinding workspaces", ginkgo.Label("rebind"), ginkgo.Ordered, func() {
		ctx := context.Background()
		initialDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		ginkgo.It("should rebind a workspace to a new provider", func() {
			tempDir, err := framework.CopyToTempDir("tests/up/testdata/no-devcontainer")
			framework.ExpectNoError(err)
			ginkgo.DeferCleanup(framework.CleanupTempDir, initialDir, tempDir)

			f := framework.NewDefaultFramework(initialDir + "/bin")

			provider1Name := "provider1-for-rebind"
			provider2Name := "provider2-for-rebind"

			// Ensure that providers are deleted
			err = f.DevPodProviderDelete(ctx, provider1Name, "--ignore-not-found")
			framework.ExpectNoError(err)
			err = f.DevPodProviderDelete(ctx, provider2Name, "--ignore-not-found")
			framework.ExpectNoError(err)

			// Add and use first provider
			err = f.DevPodProviderAdd(ctx, "docker", "--name", provider1Name)
			framework.ExpectNoError(err)
			err = f.DevPodProviderUse(ctx, provider1Name)
			framework.ExpectNoError(err)

			// Add second provider
			err = f.DevPodProviderAdd(ctx, "docker", "--name", provider2Name)
			framework.ExpectNoError(err)

			// Create workspace
			err = f.DevPodUp(ctx, tempDir)
			framework.ExpectNoError(err)

			// Rebind workspace to second provider
			err = f.DevPodWorkspaceRebind(ctx, tempDir, provider2Name)
			framework.ExpectNoError(err)

			// Verify that the workspace is now associated with the second provider
			// We can do this by trying to access it after switching to the second provider
			err = f.DevPodProviderUse(ctx, provider2Name)
			framework.ExpectNoError(err)

			_, err = f.DevPodSSH(ctx, tempDir, "echo 'hello'")
			framework.ExpectNoError(err)

			// Cleanup
			err = f.DevPodWorkspaceDelete(ctx, tempDir)
			framework.ExpectNoError(err)
			err = f.DevPodProviderDelete(ctx, provider1Name)
			framework.ExpectNoError(err)
			err = f.DevPodProviderDelete(ctx, provider2Name)
			framework.ExpectNoError(err)
		})
	})
})
