module ingress

go 1.12

replace (
	github.com/zishang520/goquery => github.com/zishang520/goquery v1.5.1
	ingress/Config => ./Config
	ingress/Set => ./Set
)

require (
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/chromedp/cdproto v0.0.0-20190526232348-edc4677101ae
	github.com/chromedp/chromedp v0.3.0
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/zishang520/goquery v0.0.0-00010101000000-000000000000
	github.com/zishang520/persistent-cookiejar v0.0.0-20190425081855-17ca2770783c
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
	gopkg.in/errgo.v1 v1.0.1 // indirect
	gopkg.in/retry.v1 v1.0.3
)
