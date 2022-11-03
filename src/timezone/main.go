package timezone

var timezones = map[int]string{
	0:      "GMT +00:00",
	3600:   "GMT +01:00",
	7200:   "GMT +02:00",
	10800:  "GMT +03:00",
	14400:  "GMT +04:00",
	18000:  "GMT +05:00",
	21600:  "GMT +06:00",
	25200:  "GMT +07:00",
	28800:  "GMT +08:00",
	32400:  "GMT +09:00",
	36000:  "GMT +10:00",
	39600:  "GMT +11:00",
	43200:  "GMT +12:00",
	46800:  "GMT +13:00",
	-3600:  "GMT -01:00",
	-7200:  "GMT -02:00",
	-10800: "GMT -03:00",
	-14400: "GMT -04:00",
	-18000: "GMT -05:00",
	-21600: "GMT -06:00",
	-25200: "GMT -07:00",
	-28800: "GMT -08:00",
	-32400: "GMT -09:00",
	-36000: "GMT -10:00",
	-39600: "GMT -11:00",
}

func GetGmtTimezone(offset int) string {
	return timezones[offset]
}
