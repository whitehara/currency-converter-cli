# currency-converter-cli
A currency converter. You can run it on CLI.

This tool uses [fawazahmed0/exchange-api](https://github.com/fawazahmed0/exchange-api) for converting the currencies.

## How to build
```
$ git clone https://github.com/whitehara/currency-converter-cli
$ cd currency-converter-cli
$ go build
```
Then, you will see `./currency-converter-cli` is built.

## How to use

### Show the latest rate

E.g. 1 USD to JPY
```
$ ./currency-converter-cli usd jpy
Rate on 2024-04-29
1 USD = 158.37325054 JPY
```
### Show the latest rates about multiple currencies

E.g. 1 USD to JPY and EUR
```
$ ./currency-converter-cli usd jpy eur
Rate on 2024-04-29
1 USD = 158.37325054 JPY
1 USD = 0.93387662 EUR
```

### Show the past rate

**The date format must be `YYYY-MM-DD`**

E.g. 1 USD to JPY on 2024-04-28
```
$ ./currency-converter-cli -date 2024-04-28 usd jpy
Rate on 2024-04-28
1 USD = 158.23406541 JPY
```

### Convert some amount of the currency

You can add amount before "FROM" currency.

E.g. convert 10 USD to JPY
```
$ ./currency-converter-cli 10 usd jpy
Rate on 2024-04-29
10 USD = 1,583.7325053999998 JPY
```

### Convert from the standard input

You can use the pipe for feeding amounts.

E.g. Feeding amounts from a list
```
$ cat list.txt
1
2
3
```
```
$ cat list.txt | ./currency-converter-cli usd jpy
Rate on 2024-04-29
1 USD = 158.37325054 JPY
2 USD = 316.74650108 JPY
3 USD = 475.11975162 JPY
```

E.g. Feeding the result of a command
``` 
$ expr 1 + 1 | ./currency-converter-cli usd jpy
Rate on 2024-04-29
2 USD = 316.74650108 JPY
```

### Print CSV format

You can use `-outcsv` option for printing CSV formant.

E.g. Feeding amounts from a list
```
$ cat list.txt
1
2
3
```
```
$ cat list.txt | ./currency-converter-cli -outcsv usd jpy
"FROM AMOUNT","FROM CURRENCY","TO AMOUNT","TO CURRENCY"
"1","USD","158.37325054","JPY"
"2","USD","316.74650108","JPY"
"3","USD","475.11975162","JPY"
```

### Check available currencies list

You can use `-l` option for printing available currencies list.
```
$ ./currency-converter-cli -l
media : Media Network
perp : Perpetual Protocol
celo : Celo
clp : Chilean Peso
kub : Bitkub Coin
(omit)
```

### Show help
You can use `-h` option for printing the help.
```
$ 
Usage:
        ./currency-converter-cli [-v] [-l] [-outcsv] [-date=YYYY-MM-DD] <NUM> FROM TO [TO...]
-v      Show version
-l      Show currency list
-outcsv Print CSV format
-date   Set rate date in YYYY-MM-DD format, default=latest
<NUM>   Amount of FROM currency, default=1
        Feeding NUMs via pipe is allowed
FROM    FROM currency
TO      TO currency, multiple currencies are allowed.
```
