// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

//TODO go test
//TODO add registration to real goes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	CLASS string = "register"
	BSTAT string = "bootstatus"
	RDCFG string = "invadercfg"
	REGIS string = "invader"
	UNREG string = "unregister"
)

type PCC struct {
	enb  bool
	ip   string
	port string
	sn   string
}

var pcc = PCC{
	enb:  false,
	ip:   "",
	port: "",
	sn:   "",
}

//FIXME remove hardcodes
func pccInit() error {
	pcc.enb = true
	pcc.ip = "192.168.101.142"
	pcc.port = "8081"
	pcc.sn = "12345678"
	return nil
}

func doPost(cmd string, msg string) (res string, err error) {
	pccURL := "http://" + pcc.ip + ":" + pcc.port + "/" + CLASS + "/" + cmd + "/" + pcc.sn
	if msg == "" {
		msg = "/" + CLASS + "/" + cmd + "/" + pcc.sn
	}

	v := url.Values{}
	v.Set("msg", msg)
	s := v.Encode()
	fmt.Printf("v.Encode(): %v\n", s)

	req, err := http.NewRequest("POST", pccURL, strings.NewReader(s))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return "", err
	}
	res = string(data)
	return
}