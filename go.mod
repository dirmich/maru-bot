module github.com/dirmich/marubot

go 1.22

// 미래 버전(Go 1.25+ 요구) 유입을 원천 차단하기 위한 강제 교체 지시어
replace (
	github.com/caarlos0/env/v11 => github.com/caarlos0/env/v11 v11.3.0
	github.com/chromedp/cdproto => github.com/chromedp/cdproto v0.0.0-20240801214329-3f85d328b335
	golang.org/x/crypto => golang.org/x/crypto v0.26.0
	golang.org/x/net => golang.org/x/net v0.28.0
	golang.org/x/sys => golang.org/x/sys v0.23.0
	modernc.org/sqlite => modernc.org/sqlite v1.31.1
)

require (
	github.com/adrianmo/go-nmea v1.10.0
	github.com/bluenviron/gomavlib/v3 v3.3.0
	github.com/bwmarrin/discordgo v0.28.1
	github.com/caarlos0/env/v11 v11.3.0
	github.com/chromedp/cdproto v0.0.0-20240801214329-3f85d328b335
	github.com/chromedp/chromedp v0.10.0
	github.com/chzyer/readline v1.5.1
	github.com/getlantern/systray v1.2.2
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/gorilla/websocket v1.5.3
	github.com/kardianos/service v1.2.4
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/slack-go/slack v0.19.0
	go.bug.st/serial v1.6.4
	golang.org/x/crypto v0.26.0
	modernc.org/sqlite v1.31.1
	periph.io/x/conn/v3 v3.7.2
	periph.io/x/host/v3 v3.8.5
)

require (
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/getlantern/context v0.0.0-20190109183933-c447772a6520 // indirect
	github.com/getlantern/errors v0.0.0-20190325191628-abdb3e3e36f7 // indirect
	github.com/getlantern/golog v0.0.0-20190830074920-4ef2e798c2d7 // indirect
	github.com/getlantern/hex v0.0.0-20190417191902-c6586a6fe0b7 // indirect
	github.com/getlantern/hidden v0.0.0-20190325191715-f02dbb02be55 // indirect
	github.com/getlantern/ops v0.0.0-20190325191751-d70cb0d6f85f // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	modernc.org/gc/v3 v3.0.0-20240107210532-573471604cb6 // indirect
	modernc.org/libc v1.55.3 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
)
