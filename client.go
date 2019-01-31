package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	xj "github.com/basgys/goxml2json"
)

type Client struct {
	client    *http.Client
	Username  string
	AuthToken string
	Host      string
}

func (c *Client) doHTTP(path, qs string, body io.Reader) (string, []*http.Cookie, error) {
	method := "GET"
	if strings.ContainsAny(qs, "set") {
		method = "POST"
	}

	uri := fmt.Sprintf("https://%s/data/%s?%s", c.Host, path, qs)
	req, _ := http.NewRequest(method, uri, body)
	req.AddCookie(&http.Cookie{
		Name:  "_appwebSessionId_",
		Value: c.AuthToken,
	})

	res, err := c.client.Do(req)
	if err != nil {
		return "", nil, errors.New("could not make request")
	}
	defer res.Body.Close()

	jsonData, err := xj.Convert(res.Body)
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return jsonData.String(), res.Cookies(), fmt.Errorf("got a non 200 status: %d", res.StatusCode)
	}

	return jsonData.String(), res.Cookies(), nil
}

func (c *Client) OpenConsole() error {
	//https://192.168.0.228/viewer.jnlp(192.168.0.228@0@root@1548390931405)
	uri := fmt.Sprintf("https://%s/viewer.jnlp(%s@0@%s@%d)", c.Host, c.Host, c.Username, time.Now().Unix())
	req, _ := http.NewRequest("GET", uri, nil)
	req.AddCookie(&http.Cookie{
		Name:  "_appwebSessionId_",
		Value: c.AuthToken,
	})
	res, err := c.client.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("bad status")
	}

	f, err := os.Create("./viewer.jnlp")
	if err != nil {
		return errors.New("could not download")
	}
	defer f.Close()

	if _, err := io.Copy(f, res.Body); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetPowerState(ps PowerState) (string, error) {
	res, _, err := c.doHTTP("", fmt.Sprintf("set=pwState:%d", ps), nil)
	return res, err
}

func (c *Client) SetBootOverride(bo BootDevice, bootOnce bool) (string, error) {
	res, _, err := c.doHTTP("", fmt.Sprintf("set=vmBootOnce:%v,firstBootDevice:%d", bootOnce, bo), nil)
	return res, err
}

func (c *Client) Query(ds ...Attribute) (string, error) {
	bufs := bytes.NewBufferString("")
	for idx, attr := range ds {
		bufs.WriteString(string(attr))
		if idx < len(ds)-1 {
			bufs.WriteString(",")
		}
	}

	res, _, err := c.doHTTP("", fmt.Sprintf("get=%s", bufs.String()), nil)
	return res, err
}

func NewFromCredentials(path string) (*Client, error) {
	credential, err := LoadCredentials(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("You should log in first")
		}
		return nil, err
	}

	c, err := NewClient(credential.Host, true)
	if err != nil {
		return nil, err
	}
	c.AuthToken = credential.AuthToken
	c.Username = credential.Username

	return c, nil
}

func NewClient(host string, skipVerify bool) (*Client, error) {
	c := &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			IdleConnTimeout: time.Second,

			TLSHandshakeTimeout: time.Second * 5,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipVerify,
			},
		},
	}
	client := &Client{client: c, Host: host}
	// err := client.Login(host, username, password)
	// if err != nil {
	// 	return nil, err
	// }

	return client, nil
}

func (c *Client) Login(username, password string) (string, error) {
	body := bytes.NewBufferString(fmt.Sprintf("user=%s&password=%s", username, password))
	_, cookies, err := c.doHTTP("login", "", body)

	if err != nil {
		return "", err
	}

	for _, cookie := range cookies {
		if cookie.Name == "_appwebSessionId_" {
			c.AuthToken = cookie.Value
			return cookie.Value, nil
		}
	}

	return "", errors.New("could not find auth token in cookie")
}

var (
	PowerStatus           = Attribute("pwState")
	SystemDescription     = Attribute("sysDesc")
	SystemRevision        = Attribute("sysRev")
	HostName              = Attribute("hostName")
	OSName                = Attribute("osName")
	OSVersion             = Attribute("osVersion")
	ServiceTag            = Attribute("svcTag")
	ExpServiceCode        = Attribute("expSvcCode")
	BiosVersion           = Attribute("biosVer")
	FirmwareVersion       = Attribute("fwVersion")
	LCCFirmwareVersion    = Attribute("LCCfwVersion")
	IPV4Enabled           = Attribute("v4Enabled")
	IPV4Address           = Attribute("v4IPAddr")
	IPV6Enabled           = Attribute("v6Enabled")
	IPV6LinkLocal         = Attribute("v6LinkLocal")
	IPV6Address           = Attribute("v6Addr")
	IPV6SiteLocal         = Attribute("v6SiteLocal")
	MacAddress            = Attribute("macAddr")
	Batteries             = Attribute("batteries")
	FanRedundancy         = Attribute("fansRedundancy")
	Fans                  = Attribute("fans")
	Intrusion             = Attribute("intrusion")
	PowerSupplyRedundancy = Attribute("psRedundancy")
	PowerSupplies         = Attribute("powerSupplies")
	RMVRedundancy         = Attribute("rmvsRedundancy")
	RemovableStorage      = Attribute("removableStorage")
	Temperatures          = Attribute("temperatures")
	Voltages              = Attribute("voltages")
	KVMEnabled            = Attribute("kvmEnabled")
	PowerBudgetData       = Attribute("budgetpowerdata")
	EventLog              = Attribute("eventLogEntries")
	BootOnce              = Attribute("vmBootOnce")
	FirstBootDevice       = Attribute("firstBootDevice")
	VFKLicense            = Attribute("vfkLicense")
	User                  = Attribute("user")
	IDRACLog              = Attribute("racLogEntries")

	PowerOff            = PowerState(0)
	PowerOn             = PowerState(1)
	ColdReboot          = PowerState(2)
	WarmReboot          = PowerState(3)
	NonMaskingInterrupt = PowerState(4)
	GracefulShutdown    = PowerState(5)

	NoOverride = BootDevice(0)
	PXE        = BootDevice(1)
	HardDrive  = BootDevice(2)
	BIOS       = BootDevice(6)
	VirtualCD  = BootDevice(8)
	LocalSD    = BootDevice(16)
	LocalCD    = BootDevice(5)
)

type Attribute string
type PowerState uint8
type BootDevice uint8
