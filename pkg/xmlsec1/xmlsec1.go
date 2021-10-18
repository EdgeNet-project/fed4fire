package xmlsec1

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os/exec"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

const Template = `
<Signature xml:id="Sig_ref0" xmlns="http://www.w3.org/2000/09/xmldsig#">
	<SignedInfo>
		<CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
		<SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
		<Reference URI="#ref0">
			<Transforms>
				<Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature" />
			</Transforms>
			<DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/>
			<DigestValue></DigestValue>
		</Reference>
	</SignedInfo>
	<SignatureValue />
	<KeyInfo>
		<X509Data>
			<X509SubjectName/>
			<X509IssuerSerial/>
			<X509Certificate/>
		</X509Data>
		<KeyValue />
	</KeyInfo>
</Signature>
`

func Sign(key rsa.PrivateKey, certificate []byte, template []byte) ([]byte, error) {
	keyFileName, err := utils.WriteTempFilePem(
		x509.MarshalPKCS1PrivateKey(&key),
		utils.PEMBlockTypeRSA,
	)
	if err != nil {
		return nil, err
	}
	defer utils.RemoveFile(keyFileName)
	certificateFileName, err := utils.WriteTempFilePem(certificate, utils.PEMBlockTypeCertificate)
	if err != nil {
		return nil, err
	}
	defer utils.RemoveFile(certificateFileName)
	templateFileName, err := utils.WriteTempFile(template)
	if err != nil {
		return nil, err
	}
	defer utils.RemoveFile(templateFileName)
	args := []string{
		"--sign",
		"--privkey-pem",
		fmt.Sprintf("%s,%s", keyFileName, certificateFileName),
		templateFileName,
	}
	cmd := exec.Command("xmlsec1", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, out)
	}
	return out, nil
}

func Verify(trustedCertificates [][]byte, document []byte) error {
	trustedFileNames, err := utils.WriteTempFilesPem(
		trustedCertificates,
		utils.PEMBlockTypeCertificate,
	)
	if err != nil {
		return err
	}
	defer utils.RemoveFiles(trustedFileNames)
	documentFileName, err := utils.WriteTempFile(document)
	if err != nil {
		return err
	}
	defer utils.RemoveFile(documentFileName)
	args := []string{"--verify"}
	for _, name := range trustedFileNames {
		args = append(args, "--trusted-pem", name)
	}
	args = append(args, documentFileName)
	cmd := exec.Command("xmlsec1", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	if !bytes.HasPrefix(out, []byte("OK")) {
		return fmt.Errorf("verification failed: %s", out)
	}
	return nil
}
