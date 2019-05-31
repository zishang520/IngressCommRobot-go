module ingress

go 1.12

replace (
	github.com/zishang520/goquery => github.com/zishang520/goquery v1.5.1
	ingress/Config => ./Config
	ingress/Set => ./Set
)
