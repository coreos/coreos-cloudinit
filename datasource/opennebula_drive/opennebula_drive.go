package opennebula_drive

import (
	"os"
	"log"
	"io/ioutil"
	b64 "encoding/base64"
	
	"github.com/SlavomirPolak/bashParser/src"
	"github.com/coreos/coreos-cloudinit/datasource"
)

type opennebulaDrive struct {
	root string
	readFile func(filename string) ([]byte, error)
	
}

func NewDatasource(root string) *opennebulaDrive {
	return &opennebulaDrive{root, ioutil.ReadFile}
}

func (ond *opennebulaDrive) IsAvailable() bool {
	_, err := os.Stat(ond.root)
	return !os.IsNotExist(err)
}

func (ond *opennebulaDrive) AvailabilityChanges() bool {
	return true
}

func (ond *opennebulaDrive) ConfigRoot() string {
	return ond.root
}

func (ond *opennebulaDrive) FetchMetadata() (metadata datasource.Metadata, err error) {
	log.Printf("Attempting to read SSH_KEY from " + ond.root + "context.sh")
	// searching for SSH_PUBLIC_KEY or SSH_KEY or PUBLIC_SSH_KEY
	val, err := fetchVariableFromShellScript(ond.root + "context.sh", "SSH_PUBLIC_KEY")
	if val == "" {
		val, err = fetchVariableFromShellScript(ond.root + "context.sh", "SSH_KEY")
		if val == "" {
			val, err = fetchVariableFromShellScript(ond.root + "context.sh", "PUBLIC_SSH_KEY")
		}
	}
	if err != nil {
		return
	}
	if val != "" {
		var sshKeyMap map[string] string
		sshKeyMap = make(map[string]string)
		sshKeyMap["SSH_KEY"] = val
		metadata.SSHPublicKeys = sshKeyMap
	}
	return 
}

func (ond *opennebulaDrive) FetchUserdata() ([]byte, error) {
	log.Printf("Attempting to read USER_DATA from " + ond.root + "context.sh")
	ret, err := fetchVariableFromShellScript(ond.root + "context.sh", "USER_DATA")
	return []byte(ret), err
}

func (ond *opennebulaDrive) Type() string {
	return "opennebula-drive"
}

func fetchVariableFromShellScript(filePath string, variableName string) (string, error) {
	variablesMap, err := bashParser.UseShlex(filePath)
	if err != nil {
		return "", err
	}
	ret := variablesMap[variableName]
	
	// checking and decoding base64
	if variableName == "USER_DATA" && variablesMap["USERDATA_ENCODING"] == "base64" {
		var err error
		ret, err = decodeBase64(ret)
		if err != nil {
			return "", err
		}
	}	
	return ret, nil
}

func decodeBase64(text string) (string, error) {
	decodedText, err := b64.StdEncoding.DecodeString(text)
	if err != nil {
		log.Printf("Error during decoding from base64.\n")
	}
	return string(decodedText), err
}