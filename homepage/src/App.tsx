import React, { useState, useEffect } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { 
  Download, 
  Shield, 
  Zap, 
  MessageSquare, 
  Globe, 
  Cpu, 
  Terminal, 
  Monitor, 
  Smartphone,
  ChevronRight,
  Sun,
  Moon,
  Github,
  CheckCircle2,
  Info
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { 
  Dialog, 
  DialogContent, 
  DialogDescription, 
  DialogHeader, 
  DialogTitle, 
  DialogTrigger,
  DialogFooter,
  DialogClose
} from "@/components/ui/dialog"
import { useTheme } from "@/context/ThemeContext"
import { useTranslation, Language } from "@/lib/i18n"
import { cn } from "@/lib/utils"
import logoImg from "./assets/logo.png"

export default function App() {
  const { theme, toggleTheme } = useTheme()
  const { language, setLanguage, t } = useTranslation()
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 20)
    window.addEventListener("scroll", handleScroll)
    return () => window.removeEventListener("scroll", handleScroll)
  }, [])

  const languages: { code: Language; label: string }[] = [
    { code: "en", label: "English" },
    { code: "ko", label: "한국어" },
    { code: "ja", label: "日本語" },
    { code: "es", label: "Español" }
  ]

  const features = [
    { icon: <Shield className="w-10 h-10 text-primary" />, title: t.feature_1_title, desc: t.feature_1_desc },
    { icon: <MessageSquare className="w-10 h-10 text-primary" />, title: t.feature_2_title, desc: t.feature_2_desc },
    { icon: <Cpu className="w-10 h-10 text-primary" />, title: t.feature_3_title, desc: t.feature_3_desc }
  ]

  const channels = [
    { id: "telegram", name: "Telegram", icon: "https://upload.wikimedia.org/wikipedia/commons/8/82/Telegram_logo.svg", howTo: t.how_to_telegram },
    { id: "discord", name: "Discord", icon: "https://upload.wikimedia.org/wikipedia/commons/6/6b/Discord_Main_Logo.svg", howTo: t.how_to_discord },
    { id: "slack", name: "Slack", icon: "https://upload.wikimedia.org/wikipedia/commons/d/d5/Slack_icon_2019.svg", howTo: t.how_to_slack },
    { id: "whatsapp", name: "WhatsApp", icon: "https://upload.wikimedia.org/wikipedia/commons/6/6b/WhatsApp.svg", howTo: t.how_to_whatsapp },
    { id: "webhook", name: "Webhook", icon: null, iconComp: <Terminal className="w-8 h-8" />, howTo: t.how_to_webhook }
  ]

  const advantages = [t.advantages_1, t.advantages_2, t.advantages_3]

  return (
    <div className="min-h-screen bg-background text-foreground selection:bg-primary/30">
      {/* Navbar */}
      <nav className={cn(
        "fixed top-0 w-full z-50 transition-all duration-300 border-b",
        scrolled ? "bg-background/80 backdrop-blur-md py-3 shadow-sm" : "bg-transparent py-5 border-transparent"
      )}>
        <div className="container mx-auto px-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center p-1 overflow-hidden">
              <img src={logoImg} alt="Marubot" className="w-full h-full object-contain" />
            </div>
            <span className="text-xl font-bold tracking-tight">MaruBot</span>
          </div>

          <div className="flex items-center gap-4">
            <div className="hidden md:flex items-center gap-2 bg-muted/50 p-1 rounded-lg border">
              {languages.map(lang => (
                <button
                  key={lang.code}
                  onClick={() => setLanguage(lang.code)}
                  className={cn(
                    "px-3 py-1 rounded-md text-xs font-medium transition-all",
                    language === lang.code ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  {lang.label}
                </button>
              ))}
            </div>

            <Button variant="ghost" size="icon" onClick={toggleTheme} className="rounded-full">
              {theme === "light" ? <Moon className="w-5 h-5" /> : <Sun className="w-5 h-5" />}
            </Button>

            <a href="https://github.com/dirmich/maru-bot" target="_blank" rel="noreferrer">
              <Button variant="outline" size="icon" className="rounded-full">
                <Github className="w-5 h-5" />
              </Button>
            </a>
          </div>
        </div>
      </nav>

      <main>
        {/* Hero Section */}
        <section className="relative pt-32 pb-20 md:pt-48 md:pb-32 overflow-hidden">
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full -z-10 overflow-hidden pointer-events-none">
            <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/10 blur-[120px] rounded-full animate-pulse" />
            <div className="absolute bottom-[10%] right-[-10%] w-[30%] h-[30%] bg-primary/5 blur-[100px] rounded-full" />
          </div>

          <div className="container mx-auto px-4 text-center">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              <h2 className="text-primary font-semibold tracking-wider uppercase text-sm mb-4">{t.hero_subtitle}</h2>
              <h1 className="text-5xl md:text-7xl font-extrabold tracking-tighter mb-6 bg-gradient-to-b from-foreground to-foreground/70 bg-clip-text text-transparent">
                {t.hero_title}
              </h1>
              <p className="text-xl text-muted-foreground max-w-2xl mx-auto mb-10 leading-relaxed">
                {t.hero_desc}
              </p>

              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button size="lg" className="rounded-full px-8 gap-2 shadow-xl shadow-primary/20">
                  <Download className="w-5 h-5" /> {t.download}
                </Button>
                <Button variant="outline" size="lg" className="rounded-full px-8 gap-2">
                   {t.get_started} <ChevronRight className="w-5 h-5" />
                </Button>
              </div>
            </motion.div>
          </div>
        </section>

        {/* Features Section */}
        <section className="py-20 bg-muted/30">
          <div className="container mx-auto px-4">
            <div className="text-center mb-16">
              <h2 className="text-3xl font-bold tracking-tight mb-4">{t.features_title}</h2>
              <div className="w-20 h-1 bg-primary mx-auto rounded-full" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {features.map((feature, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.1 }}
                >
                  <Card className="h-full border-none shadow-md hover:shadow-xl transition-all hover:-translate-y-1">
                    <CardHeader>
                      <div className="mb-4">{feature.icon}</div>
                      <CardTitle>{feature.title}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-muted-foreground">{feature.desc}</p>
                    </CardContent>
                  </Card>
                </motion.div>
              ))}
            </div>
          </div>
        </section>

        {/* Advantages/Why section */}
        <section className="py-20">
          <div className="container mx-auto px-4">
            <div className="flex flex-col md:flex-row items-center gap-12">
              <div className="flex-1">
                <h2 className="text-3xl font-bold tracking-tight mb-8">{t.advantages_title}</h2>
                <div className="space-y-4">
                  {advantages.map((adv, i) => (
                    <div key={i} className="flex items-center gap-3">
                      <CheckCircle2 className="w-6 h-6 text-primary flex-shrink-0" />
                      <span className="text-lg font-medium">{adv}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div className="flex-1 bg-muted rounded-3xl p-8 relative overflow-hidden aspect-video flex items-center justify-center border shadow-inner">
                 <Monitor className="w-32 h-32 text-primary/20 absolute -bottom-10 -right-10" />
                 <Smartphone className="w-24 h-24 text-primary/10 absolute -top-10 -left-10 rotate-12" />
                 <div className="relative text-center">
                    <Zap className="w-16 h-16 text-primary mx-auto mb-4 animate-bounce" />
                    <p className="text-2xl font-bold italic tracking-wider text-primary/40">FAST & EFFIENT</p>
                 </div>
              </div>
            </div>
          </div>
        </section>

        {/* Installation Section */}
        <section id="install" className="py-20 bg-muted/30">
          <div className="container mx-auto px-4">
            <div className="text-center mb-16">
              <h2 className="text-3xl font-bold tracking-tight mb-4">{t.install_title}</h2>
              <div className="w-20 h-1 bg-primary mx-auto rounded-full" />
            </div>

            <div className="max-w-4xl mx-auto">
              <Tabs defaultValue="windows" className="w-full">
                <TabsList className="grid w-full grid-cols-3 h-12 mb-8">
                  <TabsTrigger value="windows" className="gap-2"><Monitor className="w-4 h-4" /> {t.platform_win}</TabsTrigger>
                  <TabsTrigger value="macos" className="gap-2"><Globe className="w-4 h-4" /> {t.platform_mac}</TabsTrigger>
                  <TabsTrigger value="linux" className="gap-2"><Terminal className="w-4 h-4" /> {t.platform_linux}</TabsTrigger>
                </TabsList>
                
                <TabsContent value="windows">
                  <Card className="border-none shadow-lg">
                    <CardHeader>
                      <CardTitle>{t.win_install}</CardTitle>
                      <CardDescription>Recommended for Windows 10/11 x64.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="bg-muted p-4 rounded-lg space-y-3">
                        <p className="text-sm font-medium">{t.win_step_1}</p>
                        <p className="text-sm font-medium">{t.win_step_2}</p>
                        <p className="text-sm font-medium">{t.win_step_3}</p>
                      </div>
                      <Button className="w-full sm:w-auto rounded-full gap-2">
                        <Download className="w-4 h-4" /> marubot-windows-amd64.exe
                      </Button>
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="macos">
                   <Card className="border-none shadow-lg">
                    <CardHeader>
                      <CardTitle>{t.mac_install}</CardTitle>
                      <CardDescription>Support for Intel and Apple Silicon.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="bg-muted p-4 rounded-lg space-y-3">
                        <p className="text-sm font-medium">{t.mac_step_1}</p>
                        <p className="text-sm font-medium">{t.mac_step_2}</p>
                        <p className="text-sm font-medium">{t.mac_step_3}</p>
                      </div>
                      <div className="flex flex-col sm:flex-row gap-3">
                        <Button variant="outline" className="rounded-full gap-2 flex-1">
                          <Download className="w-4 h-4" /> Apple Silicon (.zip)
                        </Button>
                        <Button variant="outline" className="rounded-full gap-2 flex-1">
                          <Download className="w-4 h-4" /> Intel Mac (.zip)
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="linux">
                   <Card className="border-none shadow-lg">
                    <CardHeader>
                      <CardTitle>{t.linux_install}</CardTitle>
                      <CardDescription>Best for servers and IoT devices.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <p className="text-sm font-medium">{t.linux_step_1}</p>
                      <div className="bg-slate-950 text-slate-50 p-6 rounded-xl font-mono text-sm relative group overflow-x-auto shadow-2xl">
                         <div className="flex items-center gap-2 mb-2 text-slate-500 text-xs">
                            <span className="w-3 h-3 rounded-full bg-red-500/50" />
                            <span className="w-3 h-3 rounded-full bg-yellow-500/50" />
                            <span className="w-3 h-3 rounded-full bg-green-500/50" />
                         </div>
                         <code>curl -fsSL https://raw.githubusercontent.com/dirmich/maruminibot/master/install.sh | bash</code>
                         <Button size="icon" variant="ghost" className="absolute top-4 right-4 text-slate-400 hover:text-white" onClick={() => navigator.clipboard.writeText("curl -fsSL https://raw.githubusercontent.com/dirmich/maruminibot/master/install.sh | bash")}>
                            <Terminal className="w-4 h-4" />
                         </Button>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          </div>
        </section>

        {/* Channels Section */}
        <section className="py-20">
          <div className="container mx-auto px-4">
             <div className="text-center mb-16">
              <h2 className="text-3xl font-bold tracking-tight mb-4">{t.channels_title}</h2>
              <p className="text-muted-foreground">{t.tokens_title}</p>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-5 gap-6">
              {channels.map((ch) => (
                <Dialog key={ch.id}>
                  <DialogTrigger asChild>
                    <button className="flex flex-col items-center gap-3 p-6 rounded-2xl border bg-card hover:bg-accent hover:border-primary/50 transition-all group shadow-sm hover:shadow-md">
                      <div className="w-16 h-16 flex items-center justify-center grayscale group-hover:grayscale-0 transition-all duration-300">
                        {ch.icon ? <img src={ch.icon} alt={ch.name} className="w-12 h-12 object-contain" /> : ch.iconComp}
                      </div>
                      <span className="font-semibold">{ch.name}</span>
                       <div className="flex items-center gap-1 text-[10px] uppercase tracking-widest text-muted-foreground group-hover:text-primary mt-2">
                        <Info className="w-3 h-3" /> Guide
                      </div>
                    </button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                      <DialogTitle className="flex items-center gap-2">
                         {ch.icon && <img src={ch.icon} alt="" className="w-6 h-6 object-contain" />}
                         {ch.name} Guide
                      </DialogTitle>
                      <DialogDescription>
                        Follow these steps to get your integration token.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="bg-muted p-4 rounded-xl whitespace-pre-wrap text-sm leading-relaxed">
                      {ch.howTo}
                    </div>
                    <DialogFooter>
                       <DialogClose asChild>
                         <Button type="button" variant="secondary" className="rounded-full w-full">{t.close}</Button>
                       </DialogClose>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              ))}
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="py-12 border-t bg-muted/20">
        <div className="container mx-auto px-4 text-center">
           <div className="flex items-center justify-center gap-2 mb-6 text-xl font-bold tracking-tight opacity-50 grayscale hover:grayscale-0 transition-all">
            <img src={logoImg} alt="" className="w-6 h-6 object-contain" />
            <span>MaruBot</span>
          </div>
          <p className="text-sm text-muted-foreground mb-4">
            {t.footer_text}
          </p>
          <p className="text-xs text-muted-foreground/60">
            &copy; 2026 MaruBot. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  )
}
