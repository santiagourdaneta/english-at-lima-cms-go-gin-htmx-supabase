package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
    "strconv"
    "strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// --- MODELOS ---

type Sentence struct {
    ID      int `json:"id,omitempty" form:"id"`
    English string `json:"english" form:"english" binding:"required"`
    Spanish string `json:"spanish" form:"spanish" binding:"required"`
}

type Quiz struct {
    ID       int `json:"id,omitempty" form:"id"`
    Question string `json:"question" form:"question" binding:"required"`
    Option1  string `json:"opt1" form:"opt1" binding:"required"`
    Option2  string `json:"opt2" form:"opt2" binding:"required"`
    Option3  string `json:"opt3" form:"opt3" binding:"required"`
    Correct  int    `json:"correct" form:"correct" binding:"required"`
}

type Resource struct {
    ID    int `json:"id,omitempty" form:"id"`
    Title string `json:"title" form:"title" binding:"required"`
    URL   string `json:"url" form:"url" binding:"required,url"`
    Type  string `json:"type" form:"type" binding:"required"`
}

// CACHÉ EN RAM
var (
    cache    = sync.Map{}
    loginTracker   = make(map[string]*loginAttempt)
    trackerMutex   = sync.Mutex{}
    maxAttempts    = 3
    lockoutDuration = 15 * time.Minute
)

// Estructura para el Rate Limit
type loginAttempt struct {
    count     int
    blockedAt time.Time
}


// --- CLIENTE API SUPABASE (REST / Puerto 443) ---
func callSupabase(method string, table string, data interface{}) error {
    // 1. Preparar URL y Cliente (El cliente se define fuera del bucle)
    // Limpiamos el SUPABASE_URL de posibles barras finales
        base := strings.TrimSuffix(os.Getenv("SUPABASE_URL"), "/")
        
        // Si la tabla ya viene con parámetros (ej: sentences?id=eq.1), no agregamos '/' extra
        url := fmt.Sprintf("%s/rest/v1/%s", base, table)
    client := &http.Client{Timeout: 10 * time.Second}
    
    // 2. Serializar datos una sola vez
    var bodyReader io.Reader
    if data != nil {
        jsonData, _ := json.Marshal(data)
        log.Printf("DEBUG - Enviando a %s: %s", table, string(jsonData))
        bodyReader = bytes.NewBuffer(jsonData)
    }

    // 3. Sistema de Reintentos (Rate Limit)
    for i := 0; i < 3; i++ {
        req, err := http.NewRequest(method, url, bodyReader)
        if err != nil {
            return err
        }

        req.Header.Set("apikey", os.Getenv("SUPABASE_KEY"))
        req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_KEY"))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Prefer", "return=representation")

        resp, err := client.Do(req)
        if err != nil {
            log.Printf("❌ Error de red: %v", err)
            return err
        }
        
        // Importante: Siempre cerrar el cuerpo para evitar fugas de memoria
        defer resp.Body.Close()

        // Manejo de Rate Limit (429) o Bloqueos temporales (403)
        if resp.StatusCode == 429 || resp.StatusCode == 403 {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }

        // Si el status es de error (400, 401, 500...), leemos la causa
        if resp.StatusCode >= 300 {
            bodyBytes, _ := io.ReadAll(resp.Body)
            log.Printf("❌ Error Supabase (%d): %s", resp.StatusCode, string(bodyBytes))
            return fmt.Errorf("db error: %d", resp.StatusCode)
        }

        return nil // Éxito total
    }

    return fmt.Errorf("max retries reached")
}
// Función para traer datos (GET)
func getFromSupabase(table string, target interface{}, query string) error {

    base := strings.TrimSuffix(os.Getenv("SUPABASE_URL"), "/")
    url := fmt.Sprintf("%s/rest/v1/%s?%s", base, table, query)
    
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("apikey", os.Getenv("SUPABASE_KEY"))
    req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_KEY"))

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        log.Printf("❌ Fallo en GET %s: %s", table, string(bodyBytes))
        return fmt.Errorf("status: %d", resp.StatusCode)
    }

    return json.NewDecoder(resp.Body).Decode(target)
}

func main() {

    if err := godotenv.Load(); err != nil {
        log.Println("⚠️ Usando variables del sistema")
    }

    // 1. LEER VARIABLES PRIMERO
    userAdmin := os.Getenv("ADMIN_USER")
    passAdmin := os.Getenv("ADMIN_PASS")

    r := gin.Default()

    r.RedirectTrailingSlash = true
    r.RedirectFixedPath = true

    // Definir la ruta raíz 
    r.GET("/", func(c *gin.Context) {
        c.Redirect(http.StatusMovedPermanently, "public")
    })

     
    r.MaxMultipartMemory = 1 << 20 // 1MB estricto
    r.Static("/static", "./static")
    r.Static("/uploads", "./uploads")
    r.LoadHTMLGlob("templates/*")
 

    // --- MIDDLEWARE DE SEGURIDAD EXTREMA CON BLOQUEO ---

    // 2. EL MIDDLEWARE 
        authWall := func(c *gin.Context) {
            ip := c.ClientIP()
            
            // --- ESTE LOG DEBE APARECER SÍ O SÍ ---
            log.Printf(">>>> ATENCIÓN: Intento de entrada desde %s", ip)
            
            trackerMutex.Lock()
            attempt, exists := loginTracker[ip]
            if exists && attempt.count >= maxAttempts {
                if time.Since(attempt.blockedAt) < lockoutDuration {
                    trackerMutex.Unlock()
                    log.Printf("❌ BLOQUEADO TOTAL: %s", ip)
                    c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Bloqueado"})
                    return
                }
                delete(loginTracker, ip)
            }
            trackerMutex.Unlock()

            user, password, hasAuth := c.Request.BasicAuth()
            
            if !hasAuth || user != userAdmin || password != passAdmin {
                trackerMutex.Lock()
                if _, ok := loginTracker[ip]; !ok {
                    loginTracker[ip] = &loginAttempt{count: 1, blockedAt: time.Now()}
                } else {
                    loginTracker[ip].count++
                    loginTracker[ip].blockedAt = time.Now()
                }
                currCount := loginTracker[ip].count
                trackerMutex.Unlock()

                log.Printf("⚠️ Fallo #%d de %s", currCount, ip)

                c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
                c.AbortWithStatus(http.StatusUnauthorized)
                return
            }
            c.Next()
        }

    // ---------------------------------------------------------
    // 🔐 RUTA ADMINISTRATIVA (Protegida)
    // ---------------------------------------------------------
  
    // Aplicar el middleware al grupo admin 
    admin := r.Group("/admin", authWall)
    {
        admin.GET("/", func(c *gin.Context) {
            var s []Sentence
            var q []Quiz
            var rs []Resource
            getFromSupabase("sentences", &s, "order=id.desc")
            getFromSupabase("quizzes", &q, "order=id.desc")
            getFromSupabase("resources", &rs, "order=id.desc")
            c.HTML(http.StatusOK, "admin.html", gin.H{
                "Sentences": s, "Quizzes": q, "Resources": rs,
            })
        })

        admin.GET("/security/stats", func(c *gin.Context) {
            blockedCount := 0
            trackerMutex.Lock()
            for _, attempt := range loginTracker {
                if attempt.count >= maxAttempts && time.Since(attempt.blockedAt) < lockoutDuration {
                    blockedCount++
                }
            }
            trackerMutex.Unlock()

            color := "#10b981" // Verde si todo está bien
            if blockedCount > 0 { color = "#ef4444" } // Rojo si hay ataques

            c.HTML(http.StatusOK, "security-stats.html", gin.H{
                "Count": blockedCount,
                "Color": color,
            })
        })

        // --- LISTA DE IPS BLOQUEADAS (Al hacer clic en la campana) ---
        admin.GET("/security/ips", func(c *gin.Context) {
            var blockedIPs []string
            trackerMutex.Lock()
            for ip, attempt := range loginTracker {
                if attempt.count >= maxAttempts && time.Since(attempt.blockedAt) < lockoutDuration {
                    blockedIPs = append(blockedIPs, ip)
                }
            }
            trackerMutex.Unlock()
            
            c.HTML(http.StatusOK, "blocked-ips-list.html", gin.H{"IPs": blockedIPs})
        })


         // --- BUSCADOR REAL-TIME ---
         admin.GET("/search", func(c *gin.Context) {
             query := c.Query("search")
             var s []Sentence
             // Búsqueda en inglés o español simultáneamente
             filter := fmt.Sprintf("or=(english.ilike.*%s*,spanish.ilike.*%s*)", query, query)
             getFromSupabase("sentences", &s, filter)
             c.HTML(http.StatusOK, "search-results.html", gin.H{"Sentences": s})
         })


         // --- CRUD: SENTENCES ---
         admin.GET("/edit/sentence", func(c *gin.Context) {
             c.HTML(http.StatusOK, "sentence-edit-form.html", gin.H{
                 "ID": c.Query("id"), "English": c.Query("english"), "Spanish": c.Query("spanish"),
             })
         })

        admin.PATCH("/update/sentence", func(c *gin.Context) {
            idStr := c.PostForm("id")
            id, err := strconv.Atoi(idStr) // CONVERSIÓN A INT
            if err != nil {
                c.String(http.StatusBadRequest, "ID inválido")
                return
            }
            
            data := Sentence{
                ID:      id, // Ahora 'id' es int
                English: c.PostForm("english"), 
                Spanish: c.PostForm("spanish"),
            }
            
            // Pasamos el ID convertido a string para la URL de Supabase
            err = callSupabase("PATCH", "sentences?id=eq."+strconv.Itoa(id), data)
            if err != nil {
                c.String(http.StatusInternalServerError, "Error en DB")
                return
            }
            
            cache.Delete("public_data")
            c.HTML(http.StatusOK, "sentence-item.html", data)
        })

         // --- CRUD: QUIZZES ---
         admin.GET("/edit/quiz", func(c *gin.Context) {
             c.HTML(http.StatusOK, "quiz-edit-form.html", gin.H{
                 "ID": c.Query("id"), "Question": c.Query("question"),
                 "Option1": c.Query("opt1"), "Option2": c.Query("opt2"), "Option3": c.Query("opt3"),
                 "Correct": c.Query("correct"),
             })
         })

       admin.POST("/update/quiz", func(c *gin.Context) {
           var data Quiz
           
           // Bind automático: captura id, question, opt1, opt2, opt3 y correct
           if err := c.ShouldBind(&data); err != nil {
               c.String(http.StatusBadRequest, "Datos inválidos: %v", err)
               return
           }

           // Usamos data.ID que vino del formulario oculto
           err := callSupabase("PATCH", "quizzes?id=eq."+ strconv.Itoa(data.ID), data)
           if err != nil {
               log.Printf("❌ Error actualizando Quiz: %v", err)
               c.String(http.StatusInternalServerError, "Error al actualizar en la nube")
               return
           }
           
           cache.Delete("public_data")
           c.HTML(http.StatusOK, "quiz-item.html", data)
       })

         // --- EDITAR RECURSO ---
         admin.GET("/resource/edit", func(c *gin.Context) {
             c.HTML(http.StatusOK, "resource-edit-form.html", gin.H{
                 "ID":    c.Query("id"),
                 "Title": c.Query("title"),
                 "URL":   c.Query("url"),
                 "Type":  c.Query("type"),
             })
         })

         // --- UPDATE RECURSO ---
         admin.POST("/resource/update", func(c *gin.Context) {
             idStr := c.PostForm("id")
             id, err := strconv.Atoi(idStr) // CONVERSIÓN A INT
             if err != nil {
                 c.String(http.StatusBadRequest, "ID inválido")
                 return
             }

             res := Resource{
                 ID:    id,
                 Title: c.PostForm("title"),
                 Type:  c.PostForm("type"),
                 URL:   c.PostForm("old_url"),
             }

             err = callSupabase("PATCH", "resources?id=eq."+strconv.Itoa(id), res)
             if err != nil {
                 c.String(http.StatusInternalServerError, "Error en base de datos")
                 return
             }

             cache.Delete("public_data")
             c.HTML(http.StatusOK, "resource-item.html", res)
         })
         
         // --- LOGICA DE CANCELACIÓN ---
         admin.GET("/cancel/:type", func(c *gin.Context) {
             tipo := c.Param("type")
             id, _ := strconv.Atoi(c.Query("id")) // Convertir primero
             
             switch tipo {
             case "sentence":
                 c.HTML(http.StatusOK, "sentence-item.html", Sentence{
                     ID: id, English: c.Query("english"), Spanish: c.Query("spanish"),
                 })
             case "quiz":
                 c.HTML(http.StatusOK, "quiz-item.html", Quiz{
                     ID: id, Question: c.Query("question"),
                     Option1: c.Query("opt1"), Option2: c.Query("opt2"), Option3: c.Query("opt3"),
                 })
             case "resource":
                 c.HTML(http.StatusOK, "resource-item.html", Resource{
                     ID: id, Title: c.Query("title"), URL: c.Query("url"), Type: c.Query("type"),
                 })
             }
         })

         // Handlers de creación 

         admin.POST("/security/unblock-all", func(c *gin.Context) {
            trackerMutex.Lock()
            loginTracker = make(map[string]*loginAttempt) // Reset total
            trackerMutex.Unlock()
            
            log.Println("🔓 Todas las IPs han sido desbloqueadas manualmente.")
            c.HTML(http.StatusOK, "blocked-ips-list.html", gin.H{"IPs": nil})
         })

          // --- HANDLER: AGREGAR SENTENCE ---
         admin.POST("/add-sentence", func(c *gin.Context) {
             var s Sentence
             if err := c.ShouldBind(&s); err != nil {
                 c.String(http.StatusBadRequest, "Texto demasiado corto o largo")
                 return
             }

             err := callSupabase("POST", "sentences", s)
             if err != nil {
                 // Si falla la DB, mandamos un error 500. 
                 // HTMX no insertará nada en la lista si recibe un 500.
                 c.String(http.StatusInternalServerError, "No se pudo guardar en la base de datos")
                 return
             }

             // Solo llegamos aquí si Supabase aceptó el registro
             c.HTML(http.StatusOK, "sentence-item.html", s)
         })

          // --- HANDLER: AGREGAR QUIZ ---
         admin.POST("/add-quiz", func(c *gin.Context) {
              var q Quiz
              if err := c.ShouldBind(&q); err != nil {
                  c.String(http.StatusBadRequest, "Error: Verifica los campos")
                  return
              }
              err := callSupabase("POST", "quizzes", q)
              if err != nil {
                  c.String(http.StatusInternalServerError, "Error en Supabase")
                  return
              }
              cache.Delete("public_data")
              c.HTML(http.StatusOK, "quiz-item.html", q)
          })

          // --- HANDLER: AGREGAR RECURSO ---
         admin.POST("/add-resource", func(c *gin.Context) {
             var res Resource 

             // 1. Validar campos básicos del formulario
             if err := c.ShouldBind(&res); err != nil {
                 c.Status(http.StatusBadRequest)
                 return
             }

             title := c.PostForm("title")
             if len(strings.TrimSpace(title)) < 3 {
                 c.String(http.StatusBadRequest, "El título debe tener al menos 3 caracteres")
                 return
             }

             // 2. Guardar en Supabase
             err := callSupabase("POST", "resources", res)
             if err != nil {
                 c.Status(http.StatusInternalServerError)
                 return
             }

             cache.Delete("public_data")
             c.HTML(http.StatusOK, "resource-item.html", res)
         })

         admin.DELETE("/delete/:table", deleteHandler)

 }

    // ---------------------------------------------------------
    // 🌍 RUTA PÚBLICA (Solo Lectura para Alumnos)
    // ---------------------------------------------------------
 r.GET("/public", func(c *gin.Context) {
     // 1. Intentar leer de caché
     if val, found := cache.Load("public_data"); found {
         c.HTML(http.StatusOK, "index.html", val)
         return
     }

     // 2. Si no hay caché, traer datos REALES de Supabase
     var sentences []Sentence
     var quizzes []Quiz
     var resources []Resource

     // Hacemos las consultas (id=neq.0 es un truco para traer todos si el ID es numérico)
     // O simplemente usa una query vacía o "id=gt.0"
     getFromSupabase("sentences", &sentences, "select=*")
     getFromSupabase("quizzes", &quizzes, "select=*")
     getFromSupabase("resources", &resources, "select=*")

     // 3. Empaquetar datos para el template
     data := gin.H{
         "Title":     "English At Lima",
         "Sentences": sentences,
         "Quizzes":   quizzes,
         "Resources": resources,
     }

     // 4. Guardar en caché y mostrar
     cache.Store("public_data", data)
     c.HTML(http.StatusOK, "index.html", data)
 })

    r.GET("/quiz/result", func(c *gin.Context) {
        score := c.Query("score")
        total := c.Query("total")
        
        c.HTML(http.StatusOK, "quiz-result.html", gin.H{
            "Score": score,
            "Total": total,
            "ShareText": fmt.Sprintf("¡Logré %s de %s puntos en English At Lima!", score, total),
        })
    })

    log.Println("✅ English At Lima Online en puerto 8080")
    r.Run(":8080")
}

func deleteHandler(c *gin.Context) {
    table := c.Param("table")
    idStr := c.Query("id") // Viene como string de la URL (?id=123)
    
    // 1. Validar que el ID sea un número antes de proceder
    id, err := strconv.Atoi(idStr)
    if err != nil {
        log.Printf("⚠️ Intento de borrado con ID inválido: %s", idStr)
        c.Status(http.StatusBadRequest)
        return
    }

    // 2. Borrar el registro en Supabase
    // Usamos el ID convertido de nuevo a string para la URL de la API
    urlParams := fmt.Sprintf("id=eq.%d", id)
    err = callSupabase("DELETE", table+"?"+urlParams, nil)
    
    if err != nil {
        log.Printf("❌ Error al borrar en Supabase: %v", err)
        c.Status(http.StatusInternalServerError)
        return
    }

    // 3. Limpieza de caché y éxito
    cache.Delete("public_data")
    log.Printf("✅ Registro %d eliminado de la tabla %s", id, table)
    c.Status(http.StatusOK)
}