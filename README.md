# 🇬🇧 English At Lima - CMS

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white)
![HTMX](https://img.shields.io/badge/HTMX-3366CC?style=for-the-badge&logo=htmx&logoColor=white)
![Supabase](https://img.shields.io/badge/Supabase-3ECF8E?style=for-the-badge&logo=supabase&logoColor=white)

Un sistema de gestión de contenidos (CMS) ligero y ultra rápido diseñado para la enseñanza de inglés. Utiliza una arquitectura moderna sin excesos de JavaScript gracias a **HTMX** y un backend robusto en **Go**.

## 🚀 Características

- **Arquitectura SSR + HTMX:** Actualizaciones parciales de la interfaz sin recargar la página.
- **Seguridad Extrema:** Middleware de protección contra fuerza bruta con bloqueo de IP temporal.
- **Triple Validación:** Validación en frontend (HTML5), backend (Go) y base de datos (PostgreSQL Constraints).
- **Caché en RAM:** Optimización de lectura mediante `sync.Map` para las rutas públicas de los alumnos.

🗄️ Estructura de Base de Datos (Supabase)

El sistema requiere tres tablas principales:

sentences: (id, english, spanish)

quizzes: (id, question, opt1, opt2, opt3, correct)

resources: (id, title, url, type) con un Check Constraint en title (mínimo 3 caracteres).

📂 Estructura del Proyecto

/static: Archivos CSS y assets globales.

/templates: Fragmentos de HTML procesados por el motor de Go.

main.go: Lógica central, middleware y API REST.