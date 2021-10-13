package openssl

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"os/exec"
	"strings"
)

func Verify(trustedCertificates [][]byte, certificate []byte) error {
	trustedFileNames, err := utils.WriteTempFilesPem(
		trustedCertificates,
		utils.PEMBlockTypeCertificate,
	)
	if err != nil {
		return err
	}
	defer utils.RemoveFiles(trustedFileNames)
	certificateFileName, err := utils.WriteTempFilePem(certificate, utils.PEMBlockTypeCertificate)
	if err != nil {
		return err
	}
	defer utils.RemoveFile(certificateFileName)
	args := []string{"verify"}
	for _, name := range trustedFileNames {
		args = append(args, "-trusted", name)
	}
	args = append(args, certificateFileName)
	cmd := exec.Command("openssl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	if strings.TrimSpace(string(out)) != fmt.Sprintf("%s: OK", certificateFileName) {
		return fmt.Errorf("verification failed:%s", out)
	}
	return nil
}
