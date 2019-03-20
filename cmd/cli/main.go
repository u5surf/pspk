package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/sah4ez/pspk/pkg/keys"
	"github.com/sah4ez/pspk/pkg/pspk"
	"github.com/sah4ez/pspk/pkg/utils"
	"github.com/urfave/cli"
)

const (
	baseURL = "https://pspk.now.sh"
)

var (
	//Version current tools
	Version string
	// Hash revision number from git
	Hash string
	// BuildDate when building this utilitites
	BuildDate string
)

func main() {
	var (
		err error
	)

	var api pspk.PSPK
	{
		api = pspk.New(baseURL)
	}

	app := cli.NewApp()
	app.Name = "pspk"
	app.Version = Version + "." + Hash
	app.Metadata = map[string]interface{}{"builded": BuildDate}
	app.Description = "Console tool for encyption/decription data through pspk.now.sh"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "key name",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "publish",
			Usage:   `Generate x25519 pair to pspk`,
			Aliases: []string{"p"},
			Action: func(c *cli.Context) error {
				name := c.GlobalString("name")
				if name == "" {
					return fmt.Errorf("name can't be empty")
				}
				path := "./" + name

				pub, priv, err := keys.GenereateDH()
				if err != nil {
					return err
				}
				err = utils.Write(path, "pub.bin", pub[:])
				if err != nil {
					return err
				}
				err = api.Publish(name, pub[:])
				if err != nil {
					return err
				}

				err = utils.Write(path, "key.bin", priv[:])
				if err != nil {
					return err
				}

				fmt.Println("Generate key pair on x25519")
				return nil
			},
		},
		{
			Name:    "secret",
			Aliases: []string{"s"},
			Usage:   `Generate shared secret key by private and public keys from pspk by name`,
			Action: func(c *cli.Context) error {
				pubName := c.Args().Get(1)
				name := c.GlobalString("name")
				if name == "" {
					return fmt.Errorf("name can't be empty")
				}
				path := "./" + name

				priv, err := utils.Read(path, "key.bin")
				if err != nil {
					return err
				}
				pub, err := api.Load(pubName)
				if err != nil {
					return err
				}
				dh := keys.Secret(priv, pub)
				fmt.Println("secret:", base64.StdEncoding.EncodeToString(dh))

				err = utils.Write(path, "secret", dh[:])
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:    "encrypt",
			Aliases: []string{"e"},
			Usage:   `Encrypt input message with shared key`,
			Action: func(c *cli.Context) error {
				pubName := c.Args()[0]
				message := c.Args()[1:]
				name := c.GlobalString("name")
				if name == "" {
					return fmt.Errorf("name can't be empty")
				}
				path := "./" + name

				priv, err := utils.Read(path, "key.bin")
				if err != nil {
					return err
				}
				pub, err := api.Load(pubName)
				if err != nil {
					return err
				}
				chain := keys.Secret(priv, pub)

				messageKey, err := keys.LoadMaterialKey(chain)
				if err != nil {
					return err
				}

				b, err := utils.Encrypt(messageKey[64:], messageKey[:32], []byte(strings.Join(message, " ")))
				if err != nil {
					return err
				}
				fmt.Println("encrypted:", base64.StdEncoding.EncodeToString(b))
				return nil
			},
		},
		{
			Name:    "decrypt",
			Aliases: []string{"d"},
			Usage:   `Decrypt input message with shared key`,
			Action: func(c *cli.Context) error {
				pubName := c.Args().Get(0)
				message := c.Args().Get(1)
				name := c.GlobalString("name")
				if name == "" {
					return fmt.Errorf("name can't be empty")
				}
				path := "./" + name

				priv, err := utils.Read(path, "key.bin")
				if err != nil {
					return err
				}
				pub, err := api.Load(pubName)
				if err != nil {
					return err
				}
				chain := keys.Secret(priv, pub)
				messageKey, err := keys.LoadMaterialKey(chain)
				if err != nil {
					return err
				}
				bytesMessage, err := base64.StdEncoding.DecodeString(message)
				if err != nil {
					return err
				}

				b, err := utils.Decrypt(messageKey[64:], messageKey[:32], bytesMessage)
				if err != nil {
					return err
				}
				fmt.Println("decoded:", string(b))
				return nil
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println("run has error:", err.Error())
	}

}