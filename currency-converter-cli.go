package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

var Version = "0.8.0"

func main() {
	// options
	verPtr := flag.Bool("v", false, "Show version")
	listPtr := flag.Bool("l", false, "Show currency list")
	csvPtr := flag.Bool("outcsv", false, "Print CSV format")
	datetPtr := flag.String("date", "latest", "Set rate date in YYYY-MM-DD format, default=latest")

	// Print usage
	flag.Usage = func() {
		flagSet := flag.CommandLine
		fmt.Println("Usage:")
		fmt.Printf("\t%s [-v] [-l] [-outcsv] [-date=YYYY-MM-DD] <NUM> FROM TO [TO...]\n", os.Args[0])
		order := []string{"v", "l", "outcsv", "date"}
		for _, name := range order {
			flag := flagSet.Lookup(name)
			fmt.Printf("-%s\t%s\n", flag.Name, flag.Usage)
		}
		fmt.Printf("<%s>\t%s\n", "NUM", "Amount of FROM currency, default=1")
		fmt.Printf("\t%s\n", "Feeding NUMs via pipe is allowed")
		fmt.Printf("%s\t%s\n", "FROM", "FROM currency")
		fmt.Printf("%s\t%s\n", "TO", "TO currency, multiple currencies are allowed.")
	}
	flag.Parse()

	// Show version
	if *verPtr {
		fmt.Println("Version:", Version)
		os.Exit(0)
	}

	// check date
	if err := checkDate(*datetPtr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cc := &CurrencyConvert{date: *datetPtr}
	// Show currency list
	if *listPtr {
		cc.PrintCurrencyList()
		os.Exit(0)
	}
	// Exit conditions
	if len(flag.Args()) < 2 {
		fmt.Println("Not enough arguments passed")
		flag.Usage()
		os.Exit(1)
	}

	var amount float64
	var err error

	sl := &StdioLines{}

	var headerPrinted bool
	//check data from pipe are available
	if sl.checkPipeMode() {
		if checkIfArgIsNumber(getArg(0)) {
			fmt.Println("Please set FROM currency name instead of the number when you use the pipe")
			os.Exit(1)
		}
		for sl.ScanLine() {
			if checkIfArgIsNumber(sl.text) {
				amount, err = strconv.ParseFloat(sl.text, 64)
				if err != nil {
					fmt.Println(err)
				}
				if *csvPtr {
					if !headerPrinted {
						printCsvHeader()
						headerPrinted = true
					}
					cc.PrintCsvConvert(amount, getArg(0), flag.Args()[1:])
				} else {
					if !headerPrinted {
						cc.PrintRateDate(getArg(0))
						headerPrinted = true
					}
					cc.PrintConvert(amount, getArg(0), flag.Args()[1:])
				}
			}
		}
	} else {
		// check Arg is number or not
		if checkIfArgIsNumber(getArg(0)) {
			if len(flag.Args()) < 3 {
				fmt.Println("Not enough arguments passed")
				flag.Usage()
				os.Exit(1)
			}
			amount, err = strconv.ParseFloat(getArg(0), 64)
			if err != nil {
				fmt.Println(err)
			}
			if *csvPtr {
				if !headerPrinted {
					printCsvHeader()
					headerPrinted = true
				}
				cc.PrintCsvConvert(amount, getArg(1), flag.Args()[2:])
			} else {
				if !headerPrinted {
					cc.PrintRateDate(getArg(1))
					headerPrinted = true
				}
				cc.PrintConvert(amount, getArg(1), flag.Args()[2:])
			}
		} else {
			if *csvPtr {
				if !headerPrinted {
					printCsvHeader()
					headerPrinted = true
				}
				cc.PrintCsvConvert(1, getArg(0), flag.Args()[1:])
			} else {
				if !headerPrinted {
					cc.PrintRateDate(getArg(0))
					headerPrinted = true
				}
				cc.PrintConvert(1, getArg(0), flag.Args()[1:])
			}
		}
	}

}

// check date format
func checkDate(arg string) error {
	var layout string = "2006-01-02"

	// "latest" is also allowed
	if arg == "latest" {
		return nil
	}
	if _, err := time.Parse(layout, arg); err != nil {
		return err
	}
	return nil
}

// print CSV header
func printCsvHeader() {
	// print header
	fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\"\n", "FROM AMOUNT", "FROM CURRENCY", "TO AMOUNT", "TO CURRENCY")
}

// get the first argument
func getArg(arg int) string {
	var argument = flag.Args()[arg]
	argument = strings.ToUpper(argument)
	return argument
}

// check the argument is number or not
func checkIfArgIsNumber(arg string) bool {
	if _, err := strconv.ParseFloat(arg, 64); err == nil {
		return true
	}
	return false
}

// check and get data from pipe
type StdioLines struct {
	scanner *bufio.Scanner
	text    string
}

// get a scanner from STDIN
func (sl *StdioLines) getScanner() {
	sl.scanner = bufio.NewScanner(os.Stdin)
}

// check if STDIN is pipe or not
func (sl *StdioLines) checkPipeMode() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Mode()&os.ModeNamedPipe != 0
}

// scan STDIN
func (sl *StdioLines) ScanLine() bool {
	if sl.checkPipeMode() {
		if sl.scanner == nil {
			sl.getScanner()
		}
		if sl.scanner.Scan() {
			if err := sl.scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
			} else {
				sl.text = sl.scanner.Text()
				return true
			}
		}
	}
	return false
}

// Currency Convert class
type CurrencyConvert struct {
	date         string
	currencyList map[string]interface{}
	fromCurrency map[string]interface{}
}

// get currency convert rate list
func (cc *CurrencyConvert) getFromCurrency(arg string) (map[string]interface{}, error) {
	var date string

	// set "latest", if cc.date is empty
	if cc.date == "" {
		date = "latest"
	} else {
		date = cc.date
	}

	var mainUrl string = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@" + date + "/v1/currencies/"
	var subUrl string = "https://currency-api.pages.dev/npm/@fawazahmed0/currency-api@" + date + "/v1/currencies/"
	var response *http.Response
	var err error

	larg := strings.ToLower(arg)
	url := mainUrl + larg + ".min.json"
	surl := subUrl + larg + ".min.json"

	// get the response from the link
	response, err = http.Get(url)
	if err != nil {
		response, err = http.Get(surl)
		if err != nil {
			return nil, err
		}
	}
	err = json.NewDecoder(response.Body).Decode(&cc.fromCurrency)
	if err != nil {
		return nil, err
	}
	return cc.fromCurrency, nil
}

// get supported currencies
func (cc *CurrencyConvert) getCurrencyList() (map[string]interface{}, error) {
	var date string

	// set "latest", if cc.date is empty
	if cc.date == "" {
		date = "latest"
	} else {
		date = cc.date
	}

	var url string = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@" + date + "/v1/currencies.min.json"
	var surl string = "https://currency-api.pages.dev/npm/@fawazahmed0/currency-api@" + date + "/v1/currencies.min.json"
	var response *http.Response
	var err error

	// get the response from the link
	response, err = http.Get(url)
	if err != nil {
		response, err = http.Get(surl)
		if err != nil {
			return nil, err
		}
	}

	// get the the currency list
	err = json.NewDecoder(response.Body).Decode(&cc.currencyList)
	if err != nil {
		return nil, err
	}
	return cc.currencyList, nil
}

// print "from" currency rate date
func (cc *CurrencyConvert) PrintCurrencyList() {
	if len(cc.currencyList) == 0 {
		_, err := cc.getCurrencyList()
		if err != nil {
			fmt.Println(err)
		}
	}
	for name, desc := range cc.currencyList {
		fmt.Println(name, ":", desc)
	}
}

// print "from" currency rate date
func (cc *CurrencyConvert) PrintRateDate(from string) bool {
	// check "from" is currency
	if !cc.CheckIsCurrency(from) {
		return false
	}

	if len(cc.fromCurrency) == 0 {
		_, err := cc.getFromCurrency(from)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	rdate := cc.fromCurrency["date"]
	fmt.Println("Rate on", rdate)
	return true
}

// convert currency from "from" to "to"
func (cc *CurrencyConvert) ConvertCurrency(amount float64, from string, to string) (float64, error) {
	lfrom := strings.ToLower(from)
	lto := strings.ToLower(to)

	// check from and to are currencies
	if !cc.CheckIsCurrency(from) || !cc.CheckIsCurrency(to) {
		return 0, errors.New("Invalid currency name")
	}

	if len(cc.fromCurrency) == 0 {
		_, err := cc.getFromCurrency(from)
		if err != nil {
			return 0, err
		}
	}
	rtmp := cc.fromCurrency[lfrom].(map[string]interface{})
	rate, err := strconv.ParseFloat(fmt.Sprintf("%v", rtmp[lto]), 64)
	if err != nil {
		return 0, err
	}
	return rate * amount, nil
}

func (cc *CurrencyConvert) PrintConvert(amount float64, from string, arguments []string) {
	// print converted results
	var i int
	for i = 0; i < len(arguments); i++ {
		result, err := cc.ConvertCurrency(amount, from, strings.ToUpper(arguments[i]))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(humanize.Commaf(amount), from, "=", humanize.Commaf(result), strings.ToUpper(arguments[i]))
		}
	}

}

// print converted rates in CSV format
func (cc *CurrencyConvert) PrintCsvConvert(amount float64, from string, arguments []string) {
	// print converted results
	var i int
	for i = 0; i < len(arguments); i++ {
		result, err := cc.ConvertCurrency(amount, from, strings.ToUpper(arguments[i]))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\"\n", humanize.Commaf(amount), from, humanize.Commaf(result), strings.ToUpper(arguments[i]))
		}
	}
}

// check currency is valid
func (cc *CurrencyConvert) CheckIsCurrency(arg string) bool {
	larg := strings.ToLower(arg)

	// if the currency list is empty, get it
	if len(cc.currencyList) == 0 {
		_, err := cc.getCurrencyList()
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	currency := cc.currencyList[larg]
	if currency == nil {
		fmt.Println(arg, "is not a currency")
	}
	return currency != nil
}
