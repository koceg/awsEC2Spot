package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

var (
	awsProfile, zone, inst, prod string
)

type price struct {
	zone  string
	cost  float64
	count int
}

type priceSort []*price

func (p priceSort) Len() int           { return len(p) }
func (p priceSort) Less(i, j int) bool { return p[i].cost < p[j].cost }
func (p priceSort) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func init() {
	flag.StringVar(&awsProfile, "p", "default", "aws profile used")
	flag.StringVar(&zone, "z", "", "aws zone uses one in profile if not given")
	flag.StringVar(&inst, "i", "m5.large", "spot instance type")
	flag.StringVar(&prod, "d", "Linux/UNIX", "product description")
	setupFlags(flag.CommandLine)
}

func setupFlags(f *flag.FlagSet) {
	f.Usage = func() {
		fmt.Printf(
			"\nExample: %s -z eu-central-1 -i m5.large 1 #get spot price for last 24h\n"+
				"%s (-p,-z,-i,-d) <duration_in_days>\n",
			os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.CommandLine.Usage()
		os.Exit(1)
	}
	sess := newSession(awsProfile, zone)
	h := ec2.New(sess)

	i := historyInput(prod, inst)
	var historyPrice []*ec2.SpotPrice

	getSpot(h, i, &historyPrice)
	for _, s := range avg(&historyPrice) {
		fmt.Println(s.zone, s.cost)
	}
}

func newSession(profile, zone string) *session.Session {
	var sess *session.Session
	var err error
	if zone != "" {
		sess, err = session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           profile,
			Config: aws.Config{
				Region: aws.String(zone),
			},
		})
	} else {
		sess, err = session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           profile,
		})
	}
	if err != nil {
		awsError(err)
	}
	return sess
}

func awsError(err error) {
	if awsErr, ok := err.(awserr.Error); ok {
		fmt.Fprintf(os.Stderr, "Error: %s %s\n", awsErr.Code(), awsErr.Message())
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			fmt.Fprintf(os.Stderr, "ReqErr: %d %s\n", reqErr.StatusCode(),
				reqErr.RequestID())
		}
	} else {
		fmt.Println(err)
	}
	os.Exit(1)
}

func historyInput(product, instance string) *ec2.DescribeSpotPriceHistoryInput {
	d, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		awsError(err)
	}
	return &ec2.DescribeSpotPriceHistoryInput{
		StartTime: aws.Time(time.Now().AddDate(0, 0, -1*d)),
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("product-description"),
				Values: []*string{
					aws.String(product),
				},
			},
			&ec2.Filter{
				Name: aws.String("instance-type"),
				Values: []*string{
					aws.String(instance),
				},
			},
		},
	}
}

func getSpot(e *ec2.EC2, in *ec2.DescribeSpotPriceHistoryInput, sum *[]*ec2.SpotPrice) {
	result, err := e.DescribeSpotPriceHistory(in)
	if err != nil {
		awsError(err)
	}
	*sum = append(*sum, result.SpotPriceHistory...)
	if *result.NextToken != "" {
		in.NextToken = result.NextToken
		getSpot(e, in, sum)
	}
}

func avg(history *[]*ec2.SpotPrice) []*price {
	spot := make(map[string]*price)
	var avg []*price
	for _, h := range *history {
		s, _ := strconv.ParseFloat(*h.SpotPrice, 64)
		z := *h.AvailabilityZone
		if _, ok := spot[z]; !ok {
			spot[z] = new(price)
			spot[z].cost += s
			spot[z].count++
			spot[z].zone = z
			continue
		}
		spot[z].cost += s
		spot[z].count++
	}
	for _, s := range spot {
		i := len(avg)
		avg = append(avg, s)
		avg[i].cost = s.cost / float64(s.count) * 1.29
		avg[i].cost = math.Round(avg[i].cost*1000000) / 1000000
	}
	sort.Sort(priceSort(avg))
	return avg
}
