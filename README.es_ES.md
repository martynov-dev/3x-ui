# 3X-UI — fork de martynov-dev

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

Fork de [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui), un panel para administrar Xray/VPN. Su objetivo es eliminar la publicidad, respaldar el panel junto con nginx y el sitio web de cobertura de la VPN, y configurar ese sitio desde el panel.

Este fork recibe actualizaciones del repositorio original y conserva sus propias funciones. Las notas de cada lanzamiento indican la versión original incluida.

> [!IMPORTANT]
> Este proyecto está destinado únicamente al uso personal. Por favor, no lo uses para fines ilegales ni en un entorno de producción.

## Diferencias respecto al original

- Se eliminaron los anuncios de patrocinadores, su página y las solicitudes de publicidad.
- Copias completas manuales y programadas: intervalo, retención local, rutas adicionales y envío opcional a otro dispositivo por SSH.
- Edición de la configuración de nginx e `index.html`, plantilla HTTP, validación con `nginx -t` y restauración de archivos si falla la aplicación.
- La instalación y las actualizaciones usan versiones de **martynov-dev/3x-ui**. Los cambios del código original se fusionan por separado.

## Instalación y actualización

Antes de instalar, publique una versión con los binarios y las sumas de comprobación necesarios. Puede usar GitHub Actions **Release 3X-UI** o prepararlos y subirlos manualmente. La compilación continua de `main` usa `dev-latest`:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh) dev-latest
```

Para actualizar una instalación existente a este fork:

```bash
XUI_UPDATE_TAG=dev-latest bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/update.sh)
```

Guarde una copia antes de actualizar. Las etiquetas estables deben existir en este fork; no se sustituyen automáticamente por versiones del original. El canal Dev del panel también consulta este fork.

## Copias de seguridad completas

En la sección de copias y sitio web, elija el intervalo y la retención local y active la programación. Está desactivada por defecto, funciona mientras se ejecuta el panel y se comprueba cada minuto. La primera copia se ejecuta tras activarla; las tareas pendientes se reanudan después de reiniciar.

Los archivos se guardan en `/etc/x-ui/full-backups` o `$XUI_DB_FOLDER/full-backups` (directorio `0700`, archivos `0600`). Incluyen una instantánea coherente de SQLite (`database/x-ui.db`) o un volcado PostgreSQL (`database/x-ui.dump`), archivos del panel y ajustes del servicio, rutas configuradas de nginx/sitios/certificados, destinos de enlaces simbólicos y `manifest.json` con versión, fecha, fuentes y rutas ausentes. Las rutas estáticas de nginx también se detectan mediante `nginx -T`.

Es una copia de la aplicación, no una imagen del sistema operativo. Se excluyen registros, archivos anteriores y WAL/SHM activos de SQLite. Añada manualmente las rutas con variables de nginx y los archivos de otros servicios. Un error de lectura de un archivo existente cancela la copia.

Para enviar por SSH, configure servidor, puerto, usuario, directorio de destino existente y ruta de la clave privada **en el servidor del panel**. Este necesita `scp`, `ssh`, autenticación por clave y el destino verificado en `known_hosts`. El destino necesita un servidor compatible con scp y `mv` de POSIX (Linux/WSL, un router u otro servidor); no se admite una shell SSH nativa de Windows. Sin servidor remoto solo se guardan copias locales. Configure la retención remota en el destino. Los envíos incompletos terminan en `.partial`; los errores se muestran y se conserva la copia local.

Los archivos contienen contraseñas y claves privadas. SSH cifra el transporte; el archivo en disco no está cifrado.

### Restauración

En Ubuntu/Debian con SQLite, use el mismo script para instalar o restaurar como root. La copia completa de Linux debe tener la misma arquitectura y versión de distribución. El script valida el archivo y SQLite, instala dependencias, restaura el ejecutable archivado del panel, la base, nginx, el sitio y los certificados e inicia los servicios. Conserva la URL y las credenciales de la copia. --check valida sin aplicar; --replace autoriza reemplazar una instalación existente. Si falla la aplicación, se revierten los archivos; si falla la reversión, se conservan en /var/backups/x-ui-restore-*. PostgreSQL y las copias de Docker requieren restauración manual. No se restauran firewall externo, DNS ni servicios externos. Use solo archivos de confianza. El script debe publicarse en este fork antes de usar su URL.

```bash
curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh -o setup.sh

sudo bash setup.sh --backup /root/full-backup.tar.gz --check

sudo bash setup.sh --backup /root/full-backup.tar.gz
```

## Sitio web nginx

nginx debe estar instalado y ejecutándose. Seleccione una configuración detectada o una ruta dentro de `/etc/nginx` y un directorio web dentro de `/var/www` o `/usr/share/nginx/html`. Cargue los archivos existentes antes de editarlos. El editor no modifica el `nginx.conf` principal.

La plantilla HTTP escucha en el puerto 80: apunte el DNS al servidor y permita el acceso. HTTPS requiere certificados existentes y directivas adecuadas. Configure manualmente un puerto local compatible con Xray fallback. La emisión de certificados y la configuración de entradas Xray no están automatizadas.

Se validan las configuraciones actual y nueva. Si falla la escritura, validación o recarga, se restauran los archivos anteriores. Una configuración en `sites-available` necesita un enlace activo en `sites-enabled` o un `include`. Evite conflictos de dominios/puertos. Solo se edita `index.html`; los demás recursos permanecen en disco y se incluyen en la copia del directorio web.

## Integrar cambios del original

Guarde los cambios en un commit y use un árbol de trabajo limpio. Añada upstream una sola vez; después obtenga y fusione los cambios:

```bash
git remote add upstream https://github.com/MHSanaei/3x-ui.git
git fetch upstream
git switch main
git merge upstream/main
```

Resuelva conflictos, ejecute las comprobaciones del proyecto y envíe la rama al fork. La fusión conserva el historial, pero la compatibilidad con futuros cambios requiere pruebas. No restablezca la rama del fork a upstream. Actualizar el panel instalado desde una versión del fork es un paso independiente.

## Referencia del panel base

Las siguientes secciones conservan la referencia útil de la versión base. Los cambios del fork y sus comandos de instalación se describen arriba. Las imágenes de contenedor deben compilarse localmente o publicarse desde este fork; una imagen del original no incluye estos cambios.

### Características

- **Entradas multiprotocolo** — VLESS, VMess, Trojan, Shadowsocks, WireGuard, AmneziaWG, TUIC v5, Hysteria2, MTProto, HTTP, SOCKS (Mixed), Dokodemo-door / Tunnel y TUN.
- **Transportes y seguridad modernos** — TCP (Raw), mKCP, WebSocket, gRPC, HTTPUpgrade y XHTTP, protegidos con TLS, XTLS y REALITY.
- **AmneziaWG integrado** — WireGuard resistente al DPI se ejecuta dentro del panel sobre una pila de red en espacio de usuario, sin módulo del kernel, DKMS ni paquetes adicionales que instalar.
- **TUIC v5 integrado** — Proxy de alto rendimiento basado en QUIC con medición de tráfico mediante retransmisión UDP nativa, handshakes 0-RTT y control de congestión BBR.
- **Proxies MTProto** — secretos FakeTLS, ad-tags y cuotas por cliente, aplicados en caliente sin cortar las conexiones existentes.
- **Fallbacks** — sirve varios protocolos en un solo puerto (p. ej. VLESS y Trojan en el 443) usando la función de fallback de Xray.
- **Gestión por cliente** — cuotas de tráfico, fechas de caducidad, límites de IP con exenciones para direcciones de confianza, límites de dispositivos (HWID), ciclos de renovación programados, estado en línea en tiempo real y enlaces de compartición, códigos QR y suscripciones con un solo clic.
- **Estadísticas de tráfico** — por entrada, por cliente y por salida, con controles de reinicio.
- **Soporte multinodo** — gestiona y escala a través de varios servidores desde un único panel, incluida la clonación de entradas en otros nodos.
- **Salida y enrutamiento** — WARP, NordVPN, PIA, reglas de enrutamiento personalizadas, balanceadores de carga con conmutación por error entre balanceadores y encadenamiento de proxy de salida. Las categorías geosite y geoip incluidas se pueden explorar directamente desde el editor de reglas.
- **Servidor de suscripción integrado** — salida raw, JSON y Clash, seleccionada automáticamente según el User-Agent del cliente, además de [plantillas de página personalizables](docs/custom-subscription-templates.md).
- **Bots de Telegram y Discord** para monitorización y gestión remotas.
- **API RESTful** con tokens de alcance limitado y caducidad opcional, y una referencia de la API dentro del panel.
- **Panel instalable (PWA)** — ancla 3X-UI al escritorio o a la pantalla de inicio del móvil.
- **Almacenamiento flexible** — SQLite (predeterminado) o PostgreSQL.
- **14 idiomas de interfaz** con temas oscuro y claro.
- **Integración con Fail2ban** para aplicar límites de IP por cliente.

### Capturas de pantalla

<details>
<summary>Haz clic para expandir</summary>

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

### Inicio Rápido

Durante la instalación se generan un nombre de usuario, una contraseña y una ruta de acceso aleatorios. Tras la instalación, ejecuta `x-ui` para abrir el menú de gestión, donde puedes iniciar/detener el servicio, ver o restablecer tus credenciales de acceso, gestionar certificados SSL y mucho más.

Cada recurso de la publicación se publica con una suma `.sha256` junto a él. Tanto `install.sh` como el actualizador verifican el archivo contra esa suma y abortan si no coincide.

Para la documentación completa —instalación, configuración, operación y la referencia completa de la API— visita **[docs.sanaei.dev](https://docs.sanaei.dev)**.

#### Instalación desatendida

El instalador también se ejecuta de forma **no interactiva** para cloud-init.
Define `XUI_NONINTERACTIVE=1` (o canalízalo sin TTY) y realizará la instalación de principio a fin sin
ninguna pregunta, generando credenciales aleatorias y escribiéndolas en
`/etc/x-ui/install-result.env`. Consulta [`deploy/`](deploy/) para:

- [User-data de cloud-init](deploy/cloud-init/) — instalación desatendida en cualquier nube (Hetzner/AWS/DO/Vultr/GCP/Azure/Oracle)
- [Notas de Hetzner Cloud](deploy/marketplace/hetzner/) — despliegue basado en cloud-init en Hetzner

### Plataformas Compatibles

**Sistemas operativos:** Ubuntu, Debian, Armbian, Fedora, CentOS, RHEL, AlmaLinux, Rocky Linux, Oracle Linux, Amazon Linux, Virtuozzo, Arch, Manjaro, Parch, openSUSE (Tumbleweed / Leap), Alpine y Windows.

**Arquitecturas:** `amd64` · `386` · `arm64` (aarch64) · `armv7` · `armv6` · `armv5` · `s390x`.

### Opciones de Base de Datos

3X-UI admite dos backends, que se eligen durante la instalación:

- **SQLite** (predeterminado) — un único archivo en `/etc/x-ui/x-ui.db`. Sin configuración, ideal para despliegues pequeños y medianos.
- **PostgreSQL** — recomendado para un gran número de clientes o configuraciones multinodo. El instalador puede instalar PostgreSQL localmente por ti, o aceptar un DSN a un servidor existente.

En tiempo de ejecución, el backend se selecciona mediante variables de entorno (el instalador las escribe por ti en `/etc/default/x-ui`):

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

#### Migrar una instalación de SQLite existente a PostgreSQL

```bash
x-ui migrate-db --dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
# luego define XUI_DB_TYPE y XUI_DB_DSN en /etc/default/x-ui y reinicia:
systemctl restart x-ui
```

El archivo SQLite de origen permanece intacto; elimínalo manualmente una vez que hayas verificado el nuevo backend.

#### Docker

El comando predeterminado `docker compose up -d` sigue usando SQLite. Para ejecutarlo con el servicio PostgreSQL incluido, descomenta las dos líneas de variables de entorno `XUI_DB_*` en `docker-compose.yml` e inícialo con el perfil:

```bash
docker compose --profile postgres up -d
```

La imagen incluye Fail2ban (habilitado de forma predeterminada) para aplicar **límites de IP** por cliente. Fail2ban banea a los infractores con `iptables`, lo que requiere la capacidad `NET_ADMIN`. `docker-compose.yml` ya la concede mediante `cap_add`; si en su lugar inicias el contenedor con `docker run`, añade tú mismo las capacidades, de lo contrario los baneos se registran pero nunca se aplican:

```bash
docker run -d --cap-add=NET_ADMIN --cap-add=NET_RAW ... ghcr.io/martynov-dev/3x-ui
```

### Variables de Entorno

| Variable | Descripción | Predeterminado |
| --- | --- | --- |
| `XUI_DB_TYPE` | Backend de base de datos: `sqlite` o `postgres` | `sqlite` |
| `XUI_DB_DSN` | Cadena de conexión de PostgreSQL (cuando `XUI_DB_TYPE=postgres`) | — |
| `XUI_DB_FOLDER` | Directorio del archivo de base de datos SQLite | `/etc/x-ui` |
| `XUI_DB_MAX_OPEN_CONNS` | Máximo de conexiones abiertas (pool de PostgreSQL) | — |
| `XUI_DB_MAX_IDLE_CONNS` | Máximo de conexiones inactivas (pool de PostgreSQL) | — |
| `XUI_INIT_WEB_BASE_PATH` | La ruta URI inicial para el panel web | `/` |
| `XUI_ENABLE_FAIL2BAN` | Habilitar la aplicación de límites de IP basada en Fail2ban | `true` |
| `XUI_LOG_LEVEL` | Nivel de registro (`debug`, `info`, `warning`, `error`) | `info` |
| `XUI_DEBUG` | Habilitar el modo de depuración | `false` |
| `XUI_TUNNEL_HEALTH_MONITOR` | Habilitar el monitor de salud del túnel (sondea una URL y reinicia xray tras fallos repetidos; un reinicio desconecta a todos los clientes) | `false` |
| `XUI_TUNNEL_HEALTH_PROXY` | Proxy a través del cual se envía el sondeo; apúntalo a una entrada local de xray para que el sondeo pruebe el túnel (p. ej. `socks5://127.0.0.1:1080`). Vacío significa que el sondeo solo comprueba la conectividad del host | — |
| `XUI_TUNNEL_HEALTH_URL` | URL sondeada para verificar la salud del túnel | `https://www.cloudflare.com/cdn-cgi/trace` |
| `XUI_TUNNEL_HEALTH_INTERVAL` | Intervalo entre sondeos | `30s` |
| `XUI_TUNNEL_HEALTH_TIMEOUT` | Tiempo de espera por sondeo | `10s` |
| `XUI_TUNNEL_HEALTH_FAILURES` | Fallos consecutivos antes de que se active un reinicio | `3` |
| `XUI_TUNNEL_HEALTH_COOLDOWN` | Retardo mínimo entre reinicios consecutivos | `5m` |
| `NODE_TOKEN_ENCRYPTION` | Cifrado en reposo de los tokens de API de los nodos: `off`, `migration` o `required` (sin el prefijo `XUI_`) | `off` |
| `XUI_NODE_TOKEN_KEY_FILE` | Llavero JSON (modo `0600`) con el id de la clave activa y sus claves de 32 bytes en base64 | `/etc/x-ui/node_token_key.json` |
| `XUI_NODE_TOKEN_KEY` | Una única clave de 32 bytes en base64, usada solo si no se puede cargar el archivo de claves | — |

La lista completa está en la [referencia de variables de entorno](https://docs.sanaei.dev/docs/reference/env-vars).

### Idiomas Compatibles

La interfaz del panel está disponible en 14 idiomas:

English · فارسی · العربية · 中文（简体） · 中文（繁體） · Español · Русский · Українська · Türkçe · Tiếng Việt · 日本語 · Bahasa Indonesia · Português (Brasil) · Slovenčina

### Contribuir

Las contribuciones son bienvenidas. Por favor, lee la [Guía de contribución](/CONTRIBUTING.md) antes de abrir una incidencia (issue) o una solicitud de incorporación (pull request).

### Herramientas de la Comunidad

Herramientas e integraciones construidas por la comunidad alrededor de 3x-ui.

- [terraform-provider-3x-ui](https://github.com/batonogov/terraform-provider-threexui) (Licencia: **MIT**): _Gestiona inbounds, clientes, configuración del panel y configuración de Xray como código con Terraform / OpenTofu._
- [3X-UI Manager](https://github.com/yukh975/3X-UI-Manager) (Licencia: **MIT**): _Cliente nativo de Android para 3x-ui — panel de control, inbounds, clientes con compartición por QR, nodos y gestión de múltiples paneles. Disponible en F-Droid._

## Origen y licencia

Proyecto original: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Licencia: [GPL-3.0](LICENSE). Se conserva la autoría original y el nombre del módulo Go por compatibilidad. Documentación general: [docs.sanaei.dev](https://docs.sanaei.dev).
