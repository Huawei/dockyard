//adapt to rkt API
package module

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/openpgp"

	"github.com/containerops/dockyard/models"
)

//codes as below are implemented to support ACI storage
func VerifyAciSignature(acipath, signpath, pubkeyspath string) error {
	files, err := ioutil.ReadDir(pubkeyspath)
	if err != nil {
		return fmt.Errorf("Read pubkeys directory failed: %v", err.Error())
	}

	if len(files) <= 0 {
		return fmt.Errorf("No pubkey file found in %v", pubkeyspath)
	}

	var keyring openpgp.EntityList
	for _, file := range files {
		pubkeyfile, err := os.Open(pubkeyspath + "/" + file.Name())
		if err != nil {
			return err
		}
		defer pubkeyfile.Close()

		keyList, err := openpgp.ReadArmoredKeyRing(pubkeyfile)
		if err != nil {
			return err
		}

		if len(keyList) < 1 {
			return fmt.Errorf("Missing opengpg entity")
		}

		keyring = append(keyring, keyList[0])
	}

	acifile, err := os.Open(acipath)
	if err != nil {
		return fmt.Errorf("Open ACI file failed: %v", err.Error())
	}
	defer acifile.Close()

	signfile, err := os.Open(signpath)
	if err != nil {
		return fmt.Errorf("Open signature file failed: %v", err.Error())
	}
	defer signfile.Close()

	if _, err := acifile.Seek(0, 0); err != nil {
		return fmt.Errorf("Seek ACI file failed: %v", err)
	}
	if _, err := signfile.Seek(0, 0); err != nil {
		return fmt.Errorf("Seek signature file: %v", err)
	}

	//Verify detached signature which default is ASCII format
	_, err = openpgp.CheckArmoredDetachedSignature(keyring, acifile, signfile)
	if err == io.EOF {
		if _, err := acifile.Seek(0, 0); err != nil {
			return fmt.Errorf("Seek ACI file failed: %v", err)
		}
		if _, err := signfile.Seek(0, 0); err != nil {
			return fmt.Errorf("Seek signature file: %v", err)
		}

		//try to verify detached signature with binary format
		_, err = openpgp.CheckDetachedSignature(keyring, acifile, signfile)
	}
	if err == io.EOF {
		return fmt.Errorf("Signature format is invalid")
	}

	return err
}

func CheckClientStatus(reqbody []byte) error {
	clientmsg := new(models.CompleteMsg)
	if err := json.Unmarshal(reqbody, &clientmsg); err != nil {
		return fmt.Errorf("%v", err.Error())
	}

	if !clientmsg.Success {
		return fmt.Errorf("%v", clientmsg.Reason)
	}

	return nil
}

func FillRespMsg(result bool, clientreason, serverreason string) models.CompleteMsg {
	msg := models.CompleteMsg{
		Success:      result,
		Reason:       clientreason,
		ServerReason: serverreason,
	}
	return msg
}
