# SIC - Spring Initializr CLI

Interfaz de línea de comandos para generar proyectos Spring Boot basado de la página de [Spring Initializr](https://start.spring.io).

## Instalación

```bash
go install github.com/wfrs-dev/sic/cmd/sic@latest
```

## Uso

Ejecuta el comando y sigue el asistente interactivo:

```bash
sic
```

> Es necesario tener instalado Una [Nerd Font](https://www.nerdfonts.com/font-downloads) para que funcione correctamente.
> Si no se desea o no se pueda utilizar las fuentes Nerd Font, puedes utilizar la opción `-no-nerd-font`.

El formulario te guiará para configurar:

- Nombre y descripción del proyecto
- Tipo de proyecto (Maven/Gradle)
- Lenguaje (Java/Kotlin/Groovy)
- Versión de Spring Boot
- Empaquetado (Jar/War)
- Versión de Java
- Group y Artifact
- Dependencias (búsqueda interactiva con selección múltiple)

Luego de confirmar, el proyecto se descargará y extraerá en el directorio actual.

## Características

- Formulario interactivo con soporte para colores y fuentes Nerd Font
- Búsqueda y selección múltiple de dependencias
- Integración con la API v2.1 de Spring Initializr

## Licencia

MIT
