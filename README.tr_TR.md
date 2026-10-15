# 3X-UI — martynov-dev çatalı

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) Xray/VPN yönetim panelinin bir çatalıdır. Amacı reklamları kaldırmak, paneli nginx ve VPN örtü sitesiyle birlikte yedeklemek ve siteyi panelden yapılandırmaktır.

Temel sürüm **3X-UI v3.10.2**, commit [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7). Çatalın geliştirilmesine bu özgün sürümden başlanmıştır.

## Özgün projeden farkları

- Sponsor reklamları, sponsor sayfası ve reklam yükleme istekleri kaldırıldı.
- Aralık, yerel saklama sayısı, ek yollar ve başka bir cihaza isteğe bağlı SSH aktarımı ayarlanabilen elle ve zamanlanmış tam yedekler eklendi.
- nginx yapılandırması ve `index.html` düzenleyicisi, HTTP şablonu, `nginx -t` doğrulaması ve uygulama hatasında dosyaları geri alma eklendi.
- Kurulum ve panel güncellemeleri **martynov-dev/3x-ui** sürümlerini kullanır. Özgün kaynak değişiklikleri ayrı bir birleştirmeyle alınır.

## Kurulum ve güncelleme

Önce çatalın GitHub Actions ortamında **Release 3X-UI** derlemesi tamamlanmalı ve ikili dosyaları içeren bir sürüm yayımlanmalıdır. `main` dalının sürekli derlemesi `dev-latest` etiketini kullanır:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/install.sh) dev-latest
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

Tam geri yükleme elle yapılır. Tam arşivi veritabanı içe aktarma düğmesiyle yüklemeyin. `manifest.json` dosyasını inceleyin, uyumlu panel/nginx sürümlerini kurun ve paneli durdurun. `files/` içeriğini sahiplik, izin ve sembolik bağlantıları koruyarak ilgili mutlak yollara geri yükleyin. Panel durmuşken SQLite görüntüsünü veya PostgreSQL dökümünü geri yükleyin. Servisleri başlatmadan önce `nginx -t` çalıştırın. Önce ayrı bir kurulumda deneyin.

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

## Kaynak ve lisans

Özgün proje: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Lisans: [GPL-3.0](LICENSE). Özgün yazarlık ve uyumluluk için Go modülünün adı korunur. Genel belgeler: [docs.sanaei.dev](https://docs.sanaei.dev).
