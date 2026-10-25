# 3X-UI — martynov-dev çatalı

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) Xray/VPN yönetim panelinin bir çatalıdır. Amacı reklamları kaldırmak, paneli nginx ve VPN örtü sitesiyle birlikte yedeklemek ve siteyi panelden yapılandırmaktır.

Temel sürüm **3X-UI v3.10.2**, commit [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7). Çatalın geliştirilmesine bu özgün sürümden başlanmıştır.

> [!IMPORTANT]
> Bu proje yalnızca kişisel kullanım için tasarlanmıştır. Lütfen yasadışı amaçlar için veya üretim (production) ortamında kullanmayın.

## Özgün projeden farkları

- Sponsor reklamları, sponsor sayfası ve reklam yükleme istekleri kaldırıldı.
- Aralık, yerel saklama sayısı, ek yollar ve başka bir cihaza isteğe bağlı SSH aktarımı ayarlanabilen elle ve zamanlanmış tam yedekler eklendi.
- nginx yapılandırması ve `index.html` düzenleyicisi, HTTP şablonu, `nginx -t` doğrulaması ve uygulama hatasında dosyaları geri alma eklendi.
- Kurulum ve panel güncellemeleri **martynov-dev/3x-ui** sürümlerini kullanır. Özgün kaynak değişiklikleri ayrı bir birleştirmeyle alınır.

## Kurulum ve güncelleme

Kurulumdan önce gerekli ikili dosyaları ve sağlama toplamlarını içeren bir sürüm yayımlayın. GitHub Actions **Release 3X-UI** kullanabilir veya dosyaları elle hazırlayıp yükleyebilirsiniz. `main` dalının sürekli derlemesi `dev-latest` etiketini kullanır:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh) dev-latest
```

Mevcut kurulumu bu çatala güncellemek için:

```bash
XUI_UPDATE_TAG=dev-latest bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/update.sh)
```

Güncellemeden önce yedek alın. Kararlı etiketler bu çatalda bulunmalıdır; özgün sürümler otomatik kullanılmaz. Panelin Dev kanalı da bu çatalı kontrol eder.

## Tam yedekler

Yedekler ve site bölümünde aralığı ve yerel kopya sayısını seçip zamanlamayı etkinleştirin. Varsayılan olarak kapalıdır, panel çalışırken işler ve dakikada bir kontrol edilir. İlk yedek etkinleştirmeden sonra, geciken işler yeniden başlatmadan sonra çalışır.

Arşivler `/etc/x-ui/full-backups` veya `$XUI_DB_FOLDER/full-backups` altında saklanır (dizin `0700`, arşivler `0600`). Tutarlı SQLite görüntüsü (`database/x-ui.db`) veya PostgreSQL dökümü (`database/x-ui.dump`), panel dosyaları ve servis ayarları, seçilen nginx/site/sertifika yolları, sembolik bağlantı hedefleri ve sürüm, tarih, kaynaklar ile eksik yolları içeren `manifest.json` bulunur. Statik nginx yolları `nginx -T` ile de keşfedilir.

Bu, uygulama yedeğidir; işletim sistemi imajı değildir. Günlükler, eski arşivler ve etkin SQLite WAL/SHM dosyaları hariç tutulur. nginx değişkenli yolları ve diğer servis dosyalarını elle ekleyin. Var olan bir dosyanın okunamaması yedeği durdurur.

SSH aktarımı için adres, port, kullanıcı, mevcut hedef dizin ve **panel sunucusundaki** özel anahtar yolunu belirtin. Panel sunucusunda `scp`, `ssh`, anahtarla giriş ve `known_hosts` içinde doğrulanmış hedef gerekir. Hedefte scp uyumlu sunucu ve POSIX `mv` gerekir (Linux/WSL, yönlendirici veya başka bir sunucu); yerel Windows SSH kabuğu desteklenmez. Adres boşsa yalnızca yerel yedek alınır. Uzak saklama politikasını hedefte ayarlayın. Tamamlanmamış aktarımlar `.partial` ile biter; hatalar gösterilir ve yerel arşiv korunur.

Arşivlerde parolalar ve özel anahtarlar bulunur. SSH aktarımı şifreler; diskteki arşiv şifreli değildir.

### Geri yükleme

Ubuntu/Debian ve SQLite için root olarak aynı kurulum ve geri yükleme betiğini kullanın. Tam Linux yedeği aynı işlemci mimarisi ve dağıtım sürümünden olmalıdır. Betik arşivi ve SQLite görüntüsünü doğrular, bağımlılıkları kurar, arşivdeki paneli, veritabanını, nginx’i, siteyi ve sertifikaları geri yükleyip servisleri başlatır. Panel adresi ve giriş bilgileri yedekten gelir. --check uygulamadan kontrol eder; --replace mevcut kurulumun değiştirilmesine açıkça izin verir. Uygulama hatasında dosyalar geri alınır; geri alma başarısızsa eski dosyalar /var/backups/x-ui-restore-* altında korunur. PostgreSQL ve Docker yedekleri şimdilik elle geri yüklenir. Harici güvenlik duvarı, DNS ve harici servisler geri yüklenmez. Yalnızca güvenilir arşivleri kullanın. URL kullanılmadan önce betik bu çatalda yayımlanmalıdır.

```bash
curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh -o setup.sh

sudo bash setup.sh --backup /root/full-backup.tar.gz --check

sudo bash setup.sh --backup /root/full-backup.tar.gz
```

## nginx sitesi

nginx kurulu ve çalışıyor olmalıdır. Bulunan bir yapılandırmayı ya da `/etc/nginx` altında bir yol ve `/var/www` veya `/usr/share/nginx/html` altında bir site dizini seçin. Düzenlemeden önce mevcut dosyaları yükleyin. Ana `nginx.conf` değiştirilmez.

HTTP şablonu 80 portunu dinler; DNS kaydını sunucuya yönlendirin ve erişim açın. HTTPS için mevcut sertifikalar ve uygun yönergeler gerekir. Xray fallback için uyumlu yerel portu elle ayarlayın. Sertifika üretimi ve Xray inbound yapılandırması otomatik değildir.

Mevcut ve yeni yapılandırmalar doğrulanır. Yazma, doğrulama veya reload hatasında önceki dosyalar geri getirilir. `sites-available` yapılandırması için etkin `sites-enabled` bağlantısı veya `include` gerekir. Alan adı/port çakışmalarından kaçının. Yalnızca `index.html` düzenlenir; diğer kaynaklar diskte kalır ve site diziniyle yedeklenir.

## Özgün değişiklikleri birleştirme

Yerel değişiklikleri commit edin ve temiz çalışma ağacıyla başlayın. upstream uzak deposunu bir kez ekleyin, sonra değişiklikleri alıp birleştirin:

```bash
git remote add upstream https://github.com/MHSanaei/3x-ui.git
git fetch upstream
git switch main
git merge upstream/main
```

Çakışmaları çözün, proje kontrollerini çalıştırın ve dalı çatalınıza gönderin. Birleştirme commit geçmişini korur, ancak gelecekteki özgün değişikliklerle uyumluluk test gerektirir. Çatal dalını upstream’e sıfırlamayın. Çalışan paneli çatal sürümünden güncellemek ayrı bir adımdır.

## Temel panel başvurusu

Aşağıdaki bölümler temel sürümün yararlı başvuru bilgilerini korur. Çatalın değişiklikleri ve kurulum komutları yukarıdadır. Konteyner imajlarını yerel olarak derleyin veya bu çataldan yayımlayın; özgün imaj çatalın değişikliklerini içermez.

### Özellikler

- **Çoklu protokol destekli gelen bağlantılar (Inbounds)** — VLESS, VMess, Trojan, Shadowsocks, WireGuard, AmneziaWG, TUIC v5, Hysteria2, MTProto, HTTP, SOCKS (Karma), Dokodemo-door / Tunnel ve TUN.
- **Modern aktarımlar (transports) ve güvenlik** — TCP (Raw), mKCP, WebSocket, gRPC, HTTPUpgrade ve XHTTP; TLS, XTLS ve REALITY ile güvene alınmıştır.
- **Dahili AmneziaWG** — DPI'ya dayanıklı WireGuard, panelin içinde bir kullanıcı alanı ağ yığını üzerinde çalışır; çekirdek modülü, DKMS veya ek paket kurulumu gerektirmez.
- **Dahili TUIC v5** — Yerel UDP geçişi trafik ölçümü, 0-RTT el sıkışmaları ve BBR tıkanıklık kontrolü ile QUIC tabanlı yüksek performanslı proxy.
- **MTProto proxy'leri** — İstemci başına FakeTLS gizli anahtarları, reklam etiketleri (ad-tag) ve kotalar, mevcut bağlantılar kopmadan anlık olarak uygulanır.
- **Geri Dönüş (Fallbacks)** — Xray'in fallback desteğini kullanarak tek bir port üzerinde birden fazla protokole (ör. 443 üzerinde hem VLESS hem Trojan) hizmet verin.
- **Kullanıcı başına yönetim** — Trafik kotaları, bitiş tarihleri, güvenilir adreslere muafiyet tanınabilen IP sınırları, HWID cihaz sınırları, zamanlanmış yenileme döngüleri, canlı çevrimiçi (online) durumu ve tek tıkla paylaşım bağlantıları, QR kodları ve abonelikler.
- **Trafik istatistikleri** — Gelen bağlantı (Inbound), istemci ve giden bağlantı (Outbound) bazında istatistikler ve sıfırlama kontrolleri.
- **Çoklu düğüm (Multi-node) desteği** — Tek bir panel üzerinden birden fazla sunucuyu yönetin ve ölçeklendirin; gelen bağlantıları diğer düğümlere klonlayın.
- **Giden bağlantı (Outbound) ve yönlendirme** — WARP, NordVPN, PIA, özel yönlendirme kuralları, dengeleyiciler arası yük devretme destekli yük dengeleyiciler (load balancers) ve giden bağlantı proxy zincirleme (proxy chaining). Pakete dahil geosite ve geoip kategorileri doğrudan kural düzenleyicisinden taranabilir.
- **Dahili abonelik sunucusu** — İstemcinin User-Agent bilgisine göre otomatik seçilen raw, JSON ve Clash çıktısı ve [özel sayfa şablonları](docs/custom-subscription-templates.md).
- Uzaktan izleme ve yönetim için **Telegram ve Discord botları**.
- Kapsamı sınırlanmış, isteğe bağlı olarak süresi dolan token'lar ve panel içi API referansı sunan **RESTful API**.
- **Kurulabilir panel (PWA)** — 3X-UI'yi masaüstüne veya telefon ana ekranına sabitleyin.
- **Esnek depolama** — SQLite (varsayılan) veya PostgreSQL.
- Koyu ve açık tema seçenekleriyle **14 farklı UI dili**.
- Kullanıcı başına IP limitlerini zorunlu kılmak için **Fail2ban entegrasyonu**.

### Ekran Görüntüleri

<details>
<summary>Genişletmek için tıklayın</summary>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/01-overview-dark.png">
  <img alt="Genel Bakış" src="./media/01-overview-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/02-add-inbound-dark.png">
  <img alt="Gelen Bağlantılar (Inbounds)" src="./media/02-add-inbound-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/03-add-client-dark.png">
  <img alt="Kullanıcı Ekle" src="./media/03-add-client-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/05-add-nodes-dark.png">
  <img alt="Yapılandırmalar" src="./media/05-add-nodes-light.png">
</picture>

</details>

### Hızlı Başlangıç

Kurulum sırasında rastgele bir kullanıcı adı, şifre ve erişim yolu oluşturulur. Kurulumdan sonra, hizmeti başlatabileceğiniz/durdurabileceğiniz, giriş bilgilerinizi görüntüleyebileceğiniz veya sıfırlayabileceğiniz, SSL sertifikalarını yönetebileceğiniz ve çok daha fazlasını yapabileceğiniz yönetim menüsünü açmak için terminalde `x-ui` komutunu çalıştırın.

Her yayın dosyası, yanında bir `.sha256` sağlama toplamıyla birlikte yayımlanır. Hem `install.sh` hem de güncelleyici, arşivi bu toplama karşı doğrular ve uyuşmazlık halinde işlemi durdurur.

Tam dokümantasyon — kurulum, yapılandırma, işletim ve eksiksiz API referansı — için **[docs.sanaei.dev](https://docs.sanaei.dev)** adresini ziyaret edin.

#### Etkileşimsiz kurulum

Yükleyici, cloud-init için **etkileşimsiz** olarak da çalışır.
`XUI_NONINTERACTIVE=1` ayarlayın (veya TTY olmadan boru hattına aktarın); kurulum baştan
sona hiçbir soru sormadan tamamlanır, rastgele kimlik bilgileri oluşturup bunları
`/etc/x-ui/install-result.env` dosyasına yazar. Şunlar için [`deploy/`](deploy/) klasörüne bakın:

- [Cloud-init user-data](deploy/cloud-init/) — herhangi bir bulutta etkileşimsiz kurulum (Hetzner/AWS/DO/Vultr/GCP/Azure/Oracle)
- [Hetzner Cloud notları](deploy/marketplace/hetzner/) — Hetzner üzerinde cloud-init tabanlı dağıtım

### Desteklenen Platformlar

**İşletim sistemleri:** Ubuntu, Debian, Armbian, Fedora, CentOS, RHEL, AlmaLinux, Rocky Linux, Oracle Linux, Amazon Linux, Virtuozzo, Arch, Manjaro, Parch, openSUSE (Tumbleweed / Leap), Alpine ve Windows.

**Mimariler:** `amd64` · `386` · `arm64` (aarch64) · `armv7` · `armv6` · `armv5` · `s390x`.

### Veritabanı Seçenekleri

3X-UI kurulum sırasında seçilebilecek iki arka uç (backend) destekler:

- **SQLite** (varsayılan) — `/etc/x-ui/x-ui.db` konumunda tek bir dosya. Kurulum gerektirmez, küçük ve orta ölçekli dağıtımlar için idealdir.
- **PostgreSQL** — Yüksek kullanıcı sayıları veya çoklu düğüm (multi-node) kurulumları için önerilir. Yükleyici sizin için yerel olarak PostgreSQL kurabilir veya mevcut bir sunucuya DSN bağlantısı kabul edebilir.

Çalışma anında veritabanı türü ortam değişkenleri (environment variables) ile seçilir (yükleyici bunları sizin için `/etc/default/x-ui` dosyasına yazar):

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

#### Mevcut bir SQLite Kurulumunu PostgreSQL'e Taşıma

```bash
x-ui migrate-db --dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
# ardından /etc/default/x-ui içindeki XUI_DB_TYPE ve XUI_DB_DSN değerlerini ayarlayıp yeniden başlatın:
systemctl restart x-ui
```

Kaynak SQLite dosyasına dokunulmaz; yeni veritabanının düzgün çalıştığını doğruladıktan sonra eski SQLite dosyasını manuel olarak silebilirsiniz.

#### Docker

Varsayılan `docker compose up -d` komutu SQLite kullanmaya devam eder. Birlikte paketlenmiş PostgreSQL servisi ile çalıştırmak için, `docker-compose.yml` dosyasındaki iki `XUI_DB_*` değişken satırının yorumunu kaldırın ve profille başlatın:

```bash
docker compose --profile postgres up -d
```

Docker imajı, kullanıcı başına **IP limitlerini** zorunlu kılmak için Fail2ban ile (varsayılan olarak etkindir) paketlenmiştir. Fail2ban, ihlalcileri `iptables` ile engeller ve bunun için `NET_ADMIN` yetkisine ihtiyaç duyar. `docker-compose.yml` bunu zaten `cap_add` üzerinden vermektedir; ancak konteyneri bunun yerine `docker run` ile başlatırsanız bu yetkileri kendiniz eklemelisiniz, aksi takdirde yasaklamalar günlüğe kaydedilir ancak uygulanmaz:

```bash
docker run -d --cap-add=NET_ADMIN --cap-add=NET_RAW ... ghcr.io/martynov-dev/3x-ui
```

### Ortam Değişkenleri (Environment Variables)

| Değişken | Açıklama | Varsayılan |
| --- | --- | --- |
| `XUI_DB_TYPE` | Veritabanı türü: `sqlite` veya `postgres` | `sqlite` |
| `XUI_DB_DSN` | PostgreSQL bağlantı dizesi (eğer `XUI_DB_TYPE=postgres` ise) | — |
| `XUI_DB_FOLDER` | SQLite veritabanı dizini | `/etc/x-ui` |
| `XUI_DB_MAX_OPEN_CONNS` | Maksimum açık bağlantı sayısı (PostgreSQL havuzu) | — |
| `XUI_DB_MAX_IDLE_CONNS` | Maksimum boşta bekleme bağlantısı (PostgreSQL havuzu) | — |
| `XUI_INIT_WEB_BASE_PATH` | Web paneli için başlangıç URI yolu | `/` |
| `XUI_ENABLE_FAIL2BAN` | Fail2ban tabanlı IP limit uygulamasını etkinleştir | `true` |
| `XUI_LOG_LEVEL` | Günlük (Log) ayrıntı seviyesi (`debug`, `info`, `warning`, `error`) | `info` |
| `XUI_DEBUG` | Hata ayıklama (debug) modunu etkinleştir | `false` |
| `XUI_TUNNEL_HEALTH_MONITOR` | Tünel sağlık izleyicisini etkinleştir (bir URL'yi yoklar ve tekrarlanan başarısızlıklardan sonra xray'i yeniden başlatır; yeniden başlatma tüm istemcilerin bağlantısını düşürür) | `false` |
| `XUI_TUNNEL_HEALTH_PROXY` | Yoklamanın gönderildiği proxy; yoklamanın tüneli test etmesi için bunu yerel bir xray gelen bağlantısına yönlendirin (ör. `socks5://127.0.0.1:1080`). Boş bırakılırsa yoklama yalnızca ana makine bağlantısını kontrol eder | — |
| `XUI_TUNNEL_HEALTH_URL` | Tünel sağlığı için yoklanan URL | `https://www.cloudflare.com/cdn-cgi/trace` |
| `XUI_TUNNEL_HEALTH_INTERVAL` | Yoklamalar arasındaki aralık | `30s` |
| `XUI_TUNNEL_HEALTH_TIMEOUT` | Yoklama başına zaman aşımı | `10s` |
| `XUI_TUNNEL_HEALTH_FAILURES` | Yeniden başlatma tetiklenmeden önceki ardışık başarısızlık sayısı | `3` |
| `XUI_TUNNEL_HEALTH_COOLDOWN` | Ardışık yeniden başlatmalar arasındaki minimum gecikme | `5m` |
| `NODE_TOKEN_ENCRYPTION` | Düğüm API token'ları için beklemede şifreleme: `off`, `migration` veya `required` (`XUI_` öneki yoktur) | `off` |
| `XUI_NODE_TOKEN_KEY_FILE` | Etkin anahtar kimliğini ve base64 kodlu 32 baytlık anahtarlarını içeren JSON anahtarlığı (mod `0600`) | `/etc/x-ui/node_token_key.json` |
| `XUI_NODE_TOKEN_KEY` | Tek bir base64 kodlu 32 baytlık anahtar; yalnızca anahtar dosyası yüklenemediğinde kullanılır | — |

Tam liste [ortam değişkenleri referansında](https://docs.sanaei.dev/docs/reference/env-vars) yer alır.

### Desteklenen Diller

Panel arayüzü 14 farklı dilde mevcuttur:

İngilizce · Farsça · Arapça · Çince (Basitleştirilmiş) · Çince (Geleneksel) · İspanyolca · Rusça · Ukraynaca · Türkçe · Vietnamca · Japonca · Endonezce · Portekizce (Brezilya) · Slovakça

### Katkıda Bulunma

Katkılarınızı her zaman bekliyoruz. Bir sorun (issue) açmadan veya pull request (PR) göndermeden önce lütfen [Katkıda Bulunma Kılavuzunu](/CONTRIBUTING.md) okuyun.

### Topluluk Araçları

3x-ui çevresindeki topluluk tarafından oluşturulmuş araçlar ve entegrasyonlar.

- [terraform-provider-3x-ui](https://github.com/batonogov/terraform-provider-threexui) (Lisans: **MIT**): _Gelen bağlantılarnı, kullanıcıları, panel ayarlarını ve Xray yapılandırmasını Terraform / OpenTofu ile kod olarak (as code) yönetin._
- [3X-UI Manager](https://github.com/yukh975/3X-UI-Manager) (Lisans: **MIT**): _3x-ui için yerel Android istemcisi — kontrol paneli, gelen bağlantılar, QR ile paylaşımlı kullanıcılar, düğümler ve çoklu panel yönetimi. F-Droid'de mevcut._

## Kaynak ve lisans

Özgün proje: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Lisans: [GPL-3.0](LICENSE). Özgün yazarlık ve uyumluluk için Go modülünün adı korunur. Genel belgeler: [docs.sanaei.dev](https://docs.sanaei.dev).
