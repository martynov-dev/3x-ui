# 3X-UI — نسخة martynov-dev المتفرعة

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

نسخة متفرعة من [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)، لوحة إدارة Xray/VPN. هدفها إزالة الإعلانات ونسخ اللوحة احتياطياً مع nginx وموقع التمويه الخاص بـ VPN، وإعداد الموقع من داخل اللوحة.

تعتمد على **3X-UI v3.10.2**، عند الالتزام [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7). هذا هو إصدار المشروع الأصلي الذي بدأ منه تطوير هذه النسخة.

> [!IMPORTANT]
> هذا المشروع مخصص للاستخدام الشخصي فقط. يرجى عدم استخدامه لأغراض غير قانونية أو في بيئة إنتاجية.

## الاختلافات عن المشروع الأصلي

- أُزيلت إعلانات الرعاة وصفحتهم وطلبات تحميل الإعلانات.
- أُضيفت نسخ احتياطية كاملة يدوية ومجدولة، مع ضبط الفترة وعدد النسخ المحلية والمسارات الإضافية والإرسال الاختياري إلى جهاز آخر عبر SSH.
- أُضيف تحرير إعدادات nginx و`index.html` وقالب HTTP والتحقق عبر `nginx -t` واستعادة الملفات عند فشل التطبيق.
- يستخدم التثبيت وتحديث اللوحة إصدارات **martynov-dev/3x-ui**. تُدمج تغييرات الشيفرة الأصلية بشكل منفصل.

## التثبيت والتحديث

قبل التثبيت، انشر إصداراً يتضمن الملفات التنفيذية والمجاميع الاختبارية اللازمة. يمكنك استخدام GitHub Actions **Release 3X-UI** أو إعداد الملفات ورفعها يدوياً. يستخدم البناء المتجدد للفرع `main` الوسم `dev-latest`:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh) dev-latest
```

لتحديث تثبيت موجود إلى هذه النسخة:

```bash
XUI_UPDATE_TAG=dev-latest bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/update.sh)
```

احفظ نسخة احتياطية قبل التحديث. يجب نشر الوسوم المستقرة في هذا المستودع؛ لا تُستبدل تلقائياً بإصدارات المشروع الأصلي. تتحقق قناة Dev في اللوحة من هذا المستودع أيضاً.

## النسخ الاحتياطية الكاملة

في قسم النسخ الاحتياطية والموقع، اختر الفترة وعدد النسخ المحلية ثم فعّل الجدولة. تكون معطلة افتراضياً، وتعمل أثناء تشغيل اللوحة وتُفحص كل دقيقة. تُنفذ النسخة الأولى بعد التفعيل وتُستأنف المهام المتأخرة بعد إعادة التشغيل.

تُحفظ الأرشيفات في `/etc/x-ui/full-backups` أو `$XUI_DB_FOLDER/full-backups`؛ صلاحيات المجلد `0700` والأرشيف `0600`. تتضمن لقطة SQLite متسقة في `database/x-ui.db` أو تفريغ PostgreSQL في `database/x-ui.dump`، وملفات اللوحة وإعدادات الخدمة ومسارات nginx والموقع والشهادات المحددة وأهداف الروابط الرمزية و`manifest.json` الذي يسجل الإصدار والتاريخ والمصادر والمسارات المفقودة. تُكتشف أيضاً مسارات nginx الثابتة عبر `nginx -T`.

هذه نسخة للتطبيق وليست صورة لنظام التشغيل. تُستثنى السجلات والأرشيفات السابقة وملفات SQLite WAL/SHM النشطة. أضف يدوياً مسارات nginx التي تحتوي على متغيرات وملفات الخدمات الأخرى. يؤدي تعذر قراءة ملف موجود إلى إيقاف النسخ.

للإرسال عبر SSH، اضبط العنوان والمنفذ والمستخدم ومجلد الوجهة الموجود ومسار المفتاح الخاص **على خادم اللوحة**. يحتاج خادم اللوحة إلى `scp` و`ssh` والمصادقة بالمفتاح ومفتاح وجهة موثوق في `known_hosts`. تحتاج الوجهة إلى خادم متوافق مع scp وأمر POSIX `mv`، مثل Linux/WSL أو موجّه أو خادم آخر؛ لا تُدعم صدفة SSH الأصلية في Windows. ترك العنوان فارغاً يعني الحفظ المحلي فقط. تُضبط سياسة الاحتفاظ البعيد على الوجهة. تنتهي الملفات غير المكتملة بـ `.partial`؛ يظهر خطأ الإرسال وتبقى النسخة المحلية.

تحتوي الأرشيفات على كلمات مرور ومفاتيح خاصة. يشفّر SSH النقل، لكن الأرشيف نفسه غير مشفّر على القرص.

### الاستعادة

على Ubuntu/Debian مع SQLite، استخدم سكربتاً واحداً للتثبيت والاستعادة بصلاحيات root. يجب أن تكون النسخة الكاملة من Linux وبنفس معمارية المعالج وإصدار التوزيعة. يتحقق السكربت من الأرشيف وSQLite، ويثبت الاعتماديات ويستعيد ملف اللوحة التنفيذي المحفوظ وقاعدة البيانات وnginx والموقع والشهادات ويشغّل الخدمات. يبقى عنوان اللوحة وبيانات الدخول كما في النسخة. --check للفحص دون تطبيق؛ --replace يسمح صراحة باستبدال تثبيت موجود. عند فشل التطبيق تُستعاد الملفات السابقة؛ وإن فشل التراجع تبقى في /var/backups/x-ui-restore-*. تحتاج نسخ PostgreSQL وDocker حالياً إلى استعادة يدوية. لا تُستعاد قواعد الجدار الناري الخارجي أو DNS أو الخدمات الخارجية. استعد الأرشيفات الموثوقة فقط. يجب نشر السكربت في هذا المستودع قبل استخدام رابط تنزيله.

```bash
curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh -o setup.sh

sudo bash setup.sh --backup /root/full-backup.tar.gz --check

sudo bash setup.sh --backup /root/full-backup.tar.gz
```

## موقع nginx

يجب أن يكون nginx مثبتاً وقيد التشغيل. اختر إعداداً مكتشفاً أو مساراً داخل `/etc/nginx` ومجلد موقع داخل `/var/www` أو `/usr/share/nginx/html`. حمّل الملفات الموجودة قبل تحريرها. لا يعدّل المحرر ملف `nginx.conf` الرئيسي.

يستمع قالب HTTP على المنفذ 80؛ وجّه DNS إلى الخادم واسمح بالوصول. يتطلب HTTPS شهادات موجودة وتوجيهات مناسبة. اضبط يدوياً منفذاً محلياً متوافقاً مع Xray fallback. إصدار الشهادات وإعداد اتصالات Xray الواردة غير مؤتمتين.

يُتحقق من الإعدادين الحالي والجديد. يعيد فشل الكتابة أو التحقق أو reload الملفات السابقة. يحتاج إعداد `sites-available` إلى رابط نشط في `sites-enabled` أو `include`. تجنب تعارض النطاقات والمنافذ. يُحرر `index.html` فقط؛ تبقى الموارد الأخرى على القرص وتُنسخ مع مجلد الموقع.

## دمج تغييرات المشروع الأصلي

احفظ تغييراتك في التزام وابدأ بشجرة عمل نظيفة. أضف upstream مرة واحدة ثم اجلب التغييرات وادمجها:

```bash
git remote add upstream https://github.com/MHSanaei/3x-ui.git
git fetch upstream
git switch main
git merge upstream/main
```

عالج التعارضات وشغّل فحوص المشروع ثم ادفع الفرع إلى نسختك. يحافظ الدمج على سجل الالتزامات، لكن التوافق مع تغييرات مستقبلية يتطلب اختباراً. لا تُعد ضبط فرع النسخة إلى upstream. تحديث اللوحة العاملة من إصدار لهذه النسخة خطوة منفصلة.

## مرجع اللوحة الأساسية

تحتفظ الأقسام التالية بالمعلومات المفيدة من الإصدار الأساسي. اختلافات النسخة وأوامر تثبيتها موضحة أعلاه. يجب بناء صور الحاويات محلياً أو نشرها من هذا المستودع؛ صورة المشروع الأصلي لا تتضمن هذه التعديلات.

### الميزات

- **اتصالات واردة متعددة البروتوكولات** — VLESS، VMess، Trojan، Shadowsocks، WireGuard، AmneziaWG، TUIC v5، Hysteria2، MTProto، HTTP، SOCKS (Mixed)، Dokodemo-door / Tunnel و TUN.
- **وسائل نقل وأمان حديثة** — TCP (Raw)، mKCP، WebSocket، gRPC، HTTPUpgrade و XHTTP، مؤمَّنة بـ TLS و XTLS و REALITY.
- **AmneziaWG مدمج** — نسخة WireGuard المقاومة للفحص العميق للحزم (DPI) تعمل داخل اللوحة على مكدس شبكة في فضاء المستخدم، دون وحدة نواة أو DKMS أو حزم إضافية.
- **TUIC v5 مدمج** — بروكسي عالي الأداء يعتمد على QUIC مع قياس حركة المرور عبر مرحل UDP أصلي، ومصافحات 0-RTT، والتحكم في الازدحام BBR.
- **وكلاء MTProto** — أسرار FakeTLS وعلامات الإعلانات والحصص لكل عميل، تُطبَّق مباشرةً دون قطع الاتصالات القائمة.
- **Fallback** — تقديم عدة بروتوكولات على منفذ واحد (مثل VLESS و Trojan على المنفذ 443) باستخدام ميزة fallback في Xray.
- **إدارة لكل عميل** — حصص الترافيك، تواريخ انتهاء الصلاحية، حدود IP مع استثناء العناوين الموثوقة، حدود الأجهزة (HWID)، دورات تجديد مجدولة، حالة الاتصال المباشرة، وروابط مشاركة وأكواد QR واشتراكات بنقرة واحدة.
- **إحصائيات الترافيك** — لكل اتصال وارد، ولكل عميل، ولكل اتصال صادر، مع عناصر تحكم لإعادة التعيين.
- **دعم العقد المتعددة** — إدارة وتوسيع عبر عدة خوادم من لوحة واحدة، بما في ذلك استنساخ الاتصالات الواردة على عقد أخرى.
- **الاتصالات الصادرة والتوجيه** — WARP، NordVPN، PIA، قواعد توجيه مخصصة، موازنات تحميل مع تجاوز الفشل بين الموازنات، وتسلسل الوكلاء الصادرة. ويمكن تصفّح فئات geosite و geoip المضمّنة مباشرةً من محرر القواعد.
- **خادم اشتراك مدمج** — إخراج raw و JSON و Clash يُختار تلقائيًا حسب User-Agent الخاص بالعميل، مع [قوالب صفحات مخصصة](docs/custom-subscription-templates.md).
- **روبوتات تيليجرام وديسكورد** للمراقبة والإدارة عن بُعد.
- **واجهة RESTful API** مع رموز وصول محدودة النطاق وقابلة لانتهاء الصلاحية، ومرجع API داخل اللوحة.
- **لوحة قابلة للتثبيت (PWA)** — ثبّت 3X-UI على سطح المكتب أو شاشة هاتفك الرئيسية.
- **تخزين مرن** — SQLite (افتراضي) أو PostgreSQL.
- **14 لغة لواجهة المستخدم** مع سمات داكنة وفاتحة.
- **تكامل مع Fail2ban** لفرض حدود IP لكل عميل.

### لقطات الشاشة

<details>
<summary>انقر للتوسيع</summary>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/01-overview-dark.png">
  <img alt="Overview" src="./media/01-overview-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/02-add-inbound-dark.png">
  <img alt="Inbounds" src="./media/02-add-inbound-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/03-add-client-dark.png">
  <img alt="Add client" src="./media/03-add-client-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/05-add-nodes-dark.png">
  <img alt="Configs" src="./media/05-add-nodes-light.png">
</picture>

</details>

### البدء السريع

أثناء التثبيت، يتم إنشاء اسم مستخدم وكلمة مرور ومسار وصول عشوائية. بعد التثبيت، شغّل `x-ui` لفتح قائمة الإدارة، حيث يمكنك بدء/إيقاف الخدمة، وعرض أو إعادة تعيين بيانات تسجيل الدخول، وإدارة شهادات SSL، والمزيد.

يُنشر مع كل ملف إصدار مجموع تحقق `.sha256` بجانبه، ويتحقق كل من `install.sh` وأداة التحديث من الأرشيف مقابل هذا المجموع ويتوقفان عند عدم التطابق.

للحصول على الوثائق الكاملة — التثبيت والإعداد والتشغيل ومرجع API الكامل — قم بزيارة **[docs.sanaei.dev](https://docs.sanaei.dev)**.

#### التثبيت غير التفاعلي

يعمل المثبِّت أيضًا **بشكل غير تفاعلي** لـ cloud-init.
عيّن `XUI_NONINTERACTIVE=1` (أو مرّره عبر أنبوب دون TTY) وسيتولى التثبيت من البداية إلى النهاية
دون أي مطالبات، مُنشئًا بيانات اعتماد عشوائية وكاتبًا إياها في
`/etc/x-ui/install-result.env`. راجع [`deploy/`](deploy/) لـ:

- [بيانات مستخدم cloud-init](deploy/cloud-init/) — تثبيت غير تفاعلي على أي سحابة (Hetzner/AWS/DO/Vultr/GCP/Azure/Oracle)
- [ملاحظات Hetzner Cloud](deploy/marketplace/hetzner/) — نشر يعتمد على cloud-init على Hetzner

### المنصات المدعومة

**أنظمة التشغيل:** Ubuntu، Debian، Armbian، Fedora، CentOS، RHEL، AlmaLinux، Rocky Linux، Oracle Linux، Amazon Linux، Virtuozzo، Arch، Manjaro، Parch، openSUSE (Tumbleweed / Leap)، Alpine و Windows.

**المعماريات:** `amd64` · `386` · `arm64` (aarch64) · `armv7` · `armv6` · `armv5` · `s390x`.

### خيارات قاعدة البيانات

يدعم 3X-UI خلفيتين (backends) يتم اختيارهما أثناء التثبيت:

- **SQLite** (افتراضي) — ملف واحد في `/etc/x-ui/x-ui.db`. بدون إعداد، مثالي لعمليات النشر الصغيرة والمتوسطة.
- **PostgreSQL** — موصى به لأعداد العملاء الكبيرة أو الإعدادات متعددة العقد. يمكن للمثبِّت تثبيت PostgreSQL محليًا لك، أو قبول DSN لخادم موجود.

في وقت التشغيل، يتم اختيار الخلفية عبر متغيرات البيئة (يكتبها المثبِّت لك في `/etc/default/x-ui`):

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

#### ترحيل تثبيت SQLite موجود إلى PostgreSQL

```bash
x-ui migrate-db --dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
# ثم عيّن XUI_DB_TYPE و XUI_DB_DSN في /etc/default/x-ui وأعد التشغيل:
systemctl restart x-ui
```

يبقى ملف SQLite الأصلي دون تغيير؛ احذفه يدويًا بعد التحقق من الخلفية الجديدة.

#### Docker

يستمر الأمر الافتراضي `docker compose up -d` في استخدام SQLite. للتشغيل مع خدمة PostgreSQL المرفقة، أزِل التعليق عن سطري متغيرات البيئة `XUI_DB_*` في `docker-compose.yml` وشغّل باستخدام البروفايل:

```bash
docker compose --profile postgres up -d
```

تتضمن الصورة Fail2ban (مُفعَّل افتراضيًا) لفرض **حدود IP** لكل عميل. يحظر Fail2ban المخالفين باستخدام `iptables`، الذي يتطلب صلاحية `NET_ADMIN`. يمنح `docker-compose.yml` هذه الصلاحية مسبقًا عبر `cap_add`؛ إذا شغّلت الحاوية باستخدام `docker run` بدلاً من ذلك، فأضِف الصلاحيات بنفسك، وإلا فسيتم تسجيل عمليات الحظر دون تطبيقها أبدًا:

```bash
docker run -d --cap-add=NET_ADMIN --cap-add=NET_RAW ... ghcr.io/martynov-dev/3x-ui
```

### متغيرات البيئة

| المتغير | الوصف | الافتراضي |
| --- | --- | --- |
| `XUI_DB_TYPE` | خلفية قاعدة البيانات: `sqlite` أو `postgres` | `sqlite` |
| `XUI_DB_DSN` | سلسلة اتصال PostgreSQL (عندما `XUI_DB_TYPE=postgres`) | — |
| `XUI_DB_FOLDER` | مجلد ملف قاعدة بيانات SQLite | `/etc/x-ui` |
| `XUI_DB_MAX_OPEN_CONNS` | الحد الأقصى للاتصالات المفتوحة (تجمّع PostgreSQL) | — |
| `XUI_DB_MAX_IDLE_CONNS` | الحد الأقصى للاتصالات الخاملة (تجمّع PostgreSQL) | — |
| `XUI_INIT_WEB_BASE_PATH` | مسار URI الأولي للوحة الويب | `/` |
| `XUI_ENABLE_FAIL2BAN` | تفعيل فرض حدود IP المعتمد على Fail2ban | `true` |
| `XUI_LOG_LEVEL` | مستوى السجل (`debug`، `info`، `warning`، `error`) | `info` |
| `XUI_DEBUG` | تفعيل وضع التصحيح | `false` |
| `XUI_TUNNEL_HEALTH_MONITOR` | تفعيل مراقب صحة النفق (يفحص عنوان URL ويعيد تشغيل xray بعد فشل متكرر؛ إعادة التشغيل تقطع جميع العملاء) | `false` |
| `XUI_TUNNEL_HEALTH_PROXY` | الوكيل الذي يُرسَل عبره الفحص؛ وجّهه إلى اتصال xray وارد محلي ليختبر الفحص النفق (مثل `socks5://127.0.0.1:1080`). القيمة الفارغة تعني أن الفحص يتحقق فقط من اتصال المضيف | — |
| `XUI_TUNNEL_HEALTH_URL` | عنوان URL الذي يُفحَص لمعرفة صحة النفق | `https://www.cloudflare.com/cdn-cgi/trace` |
| `XUI_TUNNEL_HEALTH_INTERVAL` | الفترة بين عمليات الفحص | `30s` |
| `XUI_TUNNEL_HEALTH_TIMEOUT` | مهلة كل عملية فحص | `10s` |
| `XUI_TUNNEL_HEALTH_FAILURES` | عدد حالات الفشل المتتالية قبل تشغيل إعادة التشغيل | `3` |
| `XUI_TUNNEL_HEALTH_COOLDOWN` | الحد الأدنى للتأخير بين عمليات إعادة التشغيل المتتالية | `5m` |
| `NODE_TOKEN_ENCRYPTION` | تشفير رموز API الخاصة بالعقد أثناء التخزين: `off` أو `migration` أو `required` (بدون البادئة `XUI_`) | `off` |
| `XUI_NODE_TOKEN_KEY_FILE` | حلقة مفاتيح JSON (بأذونات `0600`) تضم معرّف المفتاح النشط ومفاتيح 32 بايت بترميز base64 | `/etc/x-ui/node_token_key.json` |
| `XUI_NODE_TOKEN_KEY` | مفتاح واحد بطول 32 بايت بترميز base64، يُستخدم فقط عند تعذّر تحميل ملف المفاتيح | — |

القائمة الكاملة متوفرة في [مرجع متغيرات البيئة](https://docs.sanaei.dev/docs/reference/env-vars).

### اللغات المدعومة

تتوفر واجهة اللوحة بـ 14 لغة:

English · فارسی · العربية · 中文（简体） · 中文（繁體） · Español · Русский · Українська · Türkçe · Tiếng Việt · 日本語 · Bahasa Indonesia · Português (Brasil) · Slovenčina

### المساهمة

المساهمات مرحب بها. يرجى قراءة [دليل المساهمة](/CONTRIBUTING.md) قبل فتح مشكلة (issue) أو طلب سحب (pull request).

### أدوات المجتمع

أدوات وتكاملات بناها المجتمع حول 3x-ui.

- [terraform-provider-3x-ui](https://github.com/batonogov/terraform-provider-threexui) (الترخيص: **MIT**): _إدارة الاتصالات الواردة والعملاء وإعدادات اللوحة وتكوين Xray كرمز باستخدام Terraform / OpenTofu._
- [3X-UI Manager](https://github.com/yukh975/3X-UI-Manager) (الترخيص: **MIT**): _عميل أندرويد أصلي لـ 3x-ui — لوحة التحكم، الاتصالات الواردة، العملاء مع مشاركة رمز QR، العقد وإدارة عدة لوحات. متاح على F-Droid._

## المصدر والترخيص

المشروع الأصلي: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). الترخيص: [GPL-3.0](LICENSE). تُحفظ نسبة العمل إلى المؤلفين الأصليين. يبقى اسم وحدة Go الأصلي لتوافق الشيفرة. التوثيق العام: [docs.sanaei.dev](https://docs.sanaei.dev).
