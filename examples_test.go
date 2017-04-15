package ini_test

import (
	"bytes"
	"fmt"
	"log"
	"os"

	ini "github.com/pierrec/go-ini"
)

func ExampleDecode() {
	buf := bytes.NewBufferString(`
[server]
host   = localhost
port   = 80
protos = http,https
`)

	type config struct {
		Host      string   `ini:"host,server"`
		Port      int      `ini:"port,server"`
		Protocols []string `ini:"protos,server"`
	}

	var conf config

	err := ini.Decode(buf, &conf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", conf)

	// Output: {Host:localhost Port:80 Protocols:[http https]}
}

func ExampleEncode() {
	buf := bytes.NewBuffer(nil)

	type config struct {
		Host      string   `ini:"host,server"`
		Port      int      `ini:"port,server"`
		Protocols []string `ini:"protos,server"`
	}

	conf := &config{"localhost", 80, []string{"http", "https"}}

	err := ini.Encode(buf, conf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", buf.Bytes())

	// Output: [server]
	// host   = localhost
	// port   = 80
	// protos = http,https
}

func ExampleIni_Decode() {
	type config struct {
		Host      string   `ini:"host,server"`
		Port      int      `ini:"port,server"`
		Protocols []string `ini:"protos,server"`
	}

	var conf config

	cini, _ := ini.New()
	cini.Set("server", "host", "localhost")
	cini.Set("server", "port", "80")
	cini.Set("server", "protos", "http,https")

	err := cini.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", conf)

	// Output: {Host:localhost Port:80 Protocols:[http https]}
}

func ExampleIni_Encode() {
	type config struct {
		Host      string   `ini:"host,server"`
		Port      int      `ini:"port,server"`
		Protocols []string `ini:"protos,server"`
	}

	conf := config{"localhost", 80, []string{"http", "https"}}

	cini, _ := ini.New()

	err := cini.Encode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	cini.WriteTo(os.Stdout)

	// Output: [server]
	// host   = localhost
	// port   = 80
	// protos = http,https
}

func ExampleIni_ReadFrom() {
	data := bytes.NewBufferString(`
id = ID2

[server]
host = localhost
port = 80
`)

	cini, _ := ini.New()

	_, err := cini.ReadFrom(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("id = %s\n", cini.Get("", "id"))
	fmt.Printf("host = %s\n", cini.Get("server", "host"))
	fmt.Printf("port = %s\n", cini.Get("server", "port"))

	// Output: id = ID2
	// host = localhost
	// port = 80
}

func ExampleIni_WriteTo() {
	cini, _ := ini.New(ini.Comment("# "))
	cini.SetComments("", "", "Main")
	cini.Set("", "id", "ID2")
	cini.SetComments("server", "", "Server definition")
	cini.Set("server", "host", "localhost")
	cini.Set("server", "port", "80")

	_, err := cini.WriteTo(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	// Output: # Main
	//
	// id = ID2
	//
	// # Server definition
	// [server]
	// host = localhost
	// port = 80
}
