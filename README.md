# Vodafone (DE) Gigacube Traffic Stats

Vodafone Germany has a product called "Gigacube" that provides LTE internet access. Currently, it's limited to 50 GB per month.

This project will scrape [the usage information page](http://quickcheck.vodafone.de/) that you can access when using your Gigacube connection. It also provides some convenience methods to estimate whether the quota will suffice for your average usage and, if not, when it will likely be used up.

I'm not trying to make this simple library a high quality product. Therefore, please look at `cmd/main.go` to see how to retrieve the data and at the `Quota.String()` method for some usage examples of the provided values.

## Example Run

```
$ go run cmd/main.go 
2018/05/12 23:24:12 32 GB of 51 GB (63%) used after 56% of the time (22167 bytes/s), estimating 6.3 GB to be missing until the end of the interval, depleted 2018-05-22 19:35:17
```
