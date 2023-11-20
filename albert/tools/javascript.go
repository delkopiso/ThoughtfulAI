//go:generate sh -c "if [ ! -f htmx.min.js ]; then curl -sLO https://unpkg.com/htmx.org@1.9.8/dist/htmx.min.js; fi"
//go:generate cp htmx.min.js ../assets/public/javascript/htmx.min.js
package tools
