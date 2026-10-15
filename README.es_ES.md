# 3X-UI — fork de martynov-dev

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

Fork de [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui), un panel para administrar Xray/VPN. Su objetivo es eliminar la publicidad, respaldar el panel junto con nginx y el sitio web de cobertura de la VPN, y configurar ese sitio desde el panel.

Basado en **3X-UI v3.10.2**, commit [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7). Esta es la versión original desde la que comenzó el desarrollo del fork.

## Diferencias respecto al original

- Se eliminaron los anuncios de patrocinadores, su página y las solicitudes de publicidad.
- Copias completas manuales y programadas: intervalo, retención local, rutas adicionales y envío opcional a otro dispositivo por SSH.
- Edición de la configuración de nginx e `index.html`, plantilla HTTP, validación con `nginx -t` y restauración de archivos si falla la aplicación.
- La instalación y las actualizaciones usan versiones de **martynov-dev/3x-ui**. Los cambios del código original se fusionan por separado.

## Instalación y actualización

Primero debe completarse la compilación **Release 3X-UI** en GitHub Actions del fork y publicarse una versión con binarios. La compilación continua de `main` usa `dev-latest`:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/install.sh) dev-latest
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

La restauración completa es manual. No use el botón de importar base de datos para un archivo completo. Revise `manifest.json`, instale versiones compatibles del panel/nginx y detenga el panel. Restaure `files/` en sus rutas absolutas conservando propietarios, permisos y enlaces simbólicos. Restaure SQLite con el panel detenido o restaure el volcado PostgreSQL. Ejecute `nginx -t` antes de iniciar los servicios. Pruebe primero en una instalación separada.

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

## Origen y licencia

Proyecto original: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Licencia: [GPL-3.0](LICENSE). Se conserva la autoría original y el nombre del módulo Go por compatibilidad. Documentación general: [docs.sanaei.dev](https://docs.sanaei.dev).
