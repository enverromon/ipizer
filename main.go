package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/route53"
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"net"
	"bytes"
	"os/user"
)

func main() {

	res, err := http.Get("http://ifcfg.net/")
	checkError(err, "Getting my ip")
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	checkError(err, "Reading the response")
	ip := net.ParseIP(string(body)).To4()
	if ip == nil {
		panic("Received data is not an ipv4 address")
	}
	println("RECEIVED:", ip.String())

	usr, _ := user.Current()
	home := usr.HomeDir
	cachePath := home + "/.ipizer/last_update"

	cache, err := ioutil.ReadFile(cachePath)
	checkError(err, "Opening the cache file")

	println("Cached data:", string(cache[:len(cache)-1]))
	last_ip := net.ParseIP(string(cache[:len(cache)-1])).To4()
	if last_ip == nil {
		panic("Cached data is not an ipv4 address")
	}
	if bytes.Compare(ip, last_ip) == 0 {
		println("The ip is valid")
		os.Exit(0)
	}

	sess := session.New(&aws.Config{
		Region: aws.String("eu-west-1"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
	r53 := route53.New(sess)

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String("home.kimia-rtb.mobi"),
						Type: aws.String("A"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip.String()),
							},
						},
						TTL: aws.Int64(300),
					},
				},
			},
		},
		HostedZoneId: aws.String("Z1LRS2ZQHM8N83"),
	}
	resp, err := r53.ChangeResourceRecordSets(params)
	checkError(err, "Updating Route53")
	fmt.Println(resp)

	err = ioutil.WriteFile(cachePath, []byte(ip.String()), 666)
	checkError(err, "Updating cache")
}


func checkError(err error, msg string) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}