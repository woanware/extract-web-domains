package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"gopkg.in/alecthomas/kingpin.v2"
	"golang.org/x/net/publicsuffix"
	"net"
	"path"
)

// ##### Constants ###########################################################

const APP_NAME string = "extract-web-domains"
const APP_VERSION string = "1.0.0"

// ##### Variables ###########################################################

var (
	inputFilePath = kingpin.Flag("input", "Input file containing the data").Short('i').Required().String()
	outputPath = kingpin.Flag("output", "Output path (directory) for results").Short('o').Required().String()
	uniqued    = kingpin.Flag("uniqued", "Output a unique list. Will hold list in memory whilst processing").Bool()
)

var (
	domains map[string]bool
	ips map[string]bool
	domain string
	line string
	indexOf int
	ipV4 net.IP
)

// ##### Methods #############################################################

// Application entry point
func main() {

	fmt.Printf("\n%s %s\n\n", APP_NAME, APP_VERSION)

	kingpin.Parse()

	inFile, err := os.Open(*inputFilePath)
	if err != nil {
		fmt.Println("Error opening input file: " + err.Error())
		return
	}
	defer inFile.Close()

	outFileDomains, err := os.Create(path.Join(*outputPath, "domains.txt"))
	if err != nil {
		fmt.Println("Error creating 'domains.txt' output file: " + err.Error())
		return
	}
	defer outFileDomains.Close()

	outFileIps, err := os.Create(path.Join(*outputPath, "ips.txt"))
	if err != nil {
		fmt.Println("Error creating 'ips.txt' output file: " + err.Error())
		return
	}
	defer outFileIps.Close()

	writerDomains := bufio.NewWriter(outFileDomains)
	writerIps := bufio.NewWriter(outFileIps)

	domains = make(map[string]bool)
	ips = make(map[string]bool)

	scanner := bufio.NewScanner(inFile)
	ret := false
	for scanner.Scan() {

		line = strings.Replace(scanner.Text(), "\"", "", -1)

		// Match against "http://"
		ret, domain = processProtocolLine(line, "http://", 7, *uniqued)
		if ret == true {
			if *uniqued == false {
				_, _ = writerDomains.WriteString(domain + "\n")
			}

			continue
		}

		// Match against "https://"
		ret, domain = processProtocolLine(line, "https://", 8, *uniqued)
		if ret == true {
			if *uniqued == false {
				_, _ = writerDomains.WriteString(domain + "\n")
			}

			continue
		}

		indexOf = strings.Index(line, "/")
		if indexOf > -1 {
			line = line[0:indexOf]
		}

		indexOf = strings.Index(line, ":")
		if indexOf > -1 {
			line = line[0:indexOf]
		}

		// Now attempt a TLD match
		ret, domain = processBasicLine(line, *uniqued)
		if ret == true {
			if *uniqued == false {
				_, _ = writerDomains.WriteString(domain + "\n")
			}

			continue
		}

		// Now match an IPv4/IPv6 address
		ret, domain = processIpLine(line, *uniqued)
		if ret == true {
			if *uniqued == false {
				_, _ = writerIps.WriteString(domain + "\n")
			}

			continue
		}
	}

	if *uniqued == true {
		for d := range domains {
			_, _ = writerDomains.WriteString(d + "\n")
		}

		for i := range ips {
			_, _ = writerIps.WriteString(i + "\n")
		}
	}

	writerDomains.Flush()
	writerIps.Flush()
}

//
func processProtocolLine(line string, protocolPrefix string, protocolPrefixLength int, uniqued bool) (bool, string){

	indexOf = strings.Index(line, protocolPrefix)
	if indexOf > -1 {
		line = line[indexOf+protocolPrefixLength:]

		indexOf = strings.Index(line, "/")
		if indexOf > -1 {
			line = line[0:indexOf]
		}

		indexOf = strings.Index(line, ":")
		if indexOf > -1 {
			line = line[0:indexOf]
		}

		if uniqued == true {
			domains[line] = true
		}

		return true, line
	}

	return false, ""
}

//
func processBasicLine(line string, uniqued bool) (bool, string){

	ps, icann := publicsuffix.PublicSuffix(line)

	if icann == false {
		if strings.Contains(ps, ".") == true {
			if uniqued == true {
				domains[line] = true
			}

			return true, line
		}
	} else {
		if uniqued == true {
			domains[line] = true
		}

		return true, line
	}

	return false, ""
}

//
func processIpLine(line string, uniqued bool) (bool, string){

	ipV4 = net.ParseIP(line)
	if ipV4.To16() != nil {
		if len(ipV4) == net.IPv4len || len(ipV4) == net.IPv6len {
			if uniqued == true {
				ips[line] = true
			}

			return true, line
		}
	}

	return false, ""
}