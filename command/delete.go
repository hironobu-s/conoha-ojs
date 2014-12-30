package command

import (
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
)

type Delete struct {
	objectName string
	*Command
}

func (cmd *Delete) parseFlags() (exitCode int, err error) {

	var showUsage bool

	fs := flag.NewFlagSet("conoha-ojs-delete", flag.ContinueOnError)
	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	if fs.NArg() < 1 {
		return ExitCodeParseFlagError, errors.New("Not enough arguments.")
	}

	// 削除するオブジェクト名
	cmd.objectName = os.Args[2]

	return ExitCodeOK, nil
}

func (cmd *Delete) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s delete <object_name> 

Delete a container or objects within a container.

<object_name> Name of object to delete.

`, lib.COMMAND_NAME)
}

func (cmd *Delete) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()

	if err != nil || exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, err
	}

	err = cmd.DeleteObject(cmd.objectName)
	if err != nil {
		return ExitCodeError, err
	}

	return ExitCodeOK, nil
}

func (cmd *Delete) DeleteObject(path string) error {
	log := lib.GetLogInstance()

	u, err := buildStorageUrl(cmd.config.EndPointUrl, path)
	if err != nil {
		return err
	}

	err = cmd.deleteObject(u)
	if err != nil {
		return err
	}
	log.Infof("%s was deleted.", path)

	return nil
}

func (cmd *Delete) deleteObject(u *url.URL) error {

	req, err := http.NewRequest(
		"DELETE",
		u.String(),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Add("X-Auth-Token", cmd.config.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	switch {
	case resp.StatusCode == 404:
		return errors.New("Object was not found.")

	// オブジェクトを含むコンテナを削除すると409 Conflictになる
	case resp.StatusCode == 409:
		return errors.New("Server returned 409 error code. (Did you try to delete the container containing objects?)")

	case resp.StatusCode >= 400:
		return errors.New("Return error code from Server.")
	}

	return nil
}
