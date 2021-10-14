package openssl

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

// TODO: Allow to verify certificate chain.
// VerifyChain? or rename certificate to certificateChain?
func Verify(trustedCertificates [][]byte, certificateChain [][]byte) error {
	trustedFileNames, err := utils.WriteTempFilesPem(
		trustedCertificates,
		utils.PEMBlockTypeCertificate,
	)
	if err != nil {
		return err
	}
	defer utils.RemoveFiles(trustedFileNames)
	certificateChainFilename, err := utils.WriteTempFilePems(
		certificateChain,
		utils.PEMBlockTypeCertificate,
	)
	if err != nil {
		return err
	}
	defer utils.RemoveFile(certificateChainFilename)
	args := []string{"verify"}
	for _, name := range trustedFileNames {
		args = append(args, "-trusted", name)
	}
	args = append(args, "-untrusted", certificateChainFilename, certificateChainFilename)
	cmd := exec.Command("openssl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	if strings.TrimSpace(string(out)) != fmt.Sprintf("%s: OK", certificateChainFilename) {
		return fmt.Errorf("verification failed: %s", out)
	}
	return nil
}
