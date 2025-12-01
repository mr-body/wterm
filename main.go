package main

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "flag"
    "log"
    "net/http"

    "github.com/creack/pty"
    "github.com/gorilla/websocket"
    "os"
    "os/exec"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

const serverSecret = "CHAVE-MUITO-SECRETA-ALTERAR-AQUI" // coloque algo GRANDE e forte

// --- cria token aleatório seguro ---
func generateToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

// --- assinatura HMAC do token ---
func signToken(token string) string {
    mac := hmac.New(sha256.New, []byte(serverSecret))
    mac.Write([]byte(token))
    return hex.EncodeToString(mac.Sum(nil))
}

// --- endpoint seguro para criar token ---
func authHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*") // permite qualquer origem
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    
    const password = "123" // senha verdadeira só usada aqui!

    if r.Method != "POST" {
        http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
        return
    }

    if r.FormValue("password") != password {
        http.Error(w, "Senha incorreta", http.StatusUnauthorized)
        return
    }

    token, err := generateToken()
    if err != nil {
        http.Error(w, "Erro interno", http.StatusInternalServerError)
        return
    }

    signature := signToken(token)

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"token":"` + token + `","signature":"` + signature + `"}`))
}

// --- valida token + assinatura ---
func validateToken(token, signature string) bool {
    expected := signToken(token)
    return hmac.Equal([]byte(signature), []byte(expected))
}

func handleWS(w http.ResponseWriter, r *http.Request) {

    token := r.URL.Query().Get("token")
    signature := r.URL.Query().Get("signature")

    if !validateToken(token, signature) {
        http.Error(w, "Acesso negado", http.StatusUnauthorized)
        return
    }

    // upgrade
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Erro:", err)
        return
    }
    defer ws.Close()

    shell := os.Getenv("SHELL")
    if shell == "" {
        shell = "/bin/bash"
    }

    cmd := exec.Command(shell)

    // força o shell a começar no diretório HOME
    cmd.Dir = os.Getenv("HOME")
    
    cmd.Env = append(os.Environ(), "TERM=xterm-256color", "COLORTERM=truecolor")

    ptyFile, err := pty.Start(cmd)
    if err != nil {
        log.Println("Erro ao iniciar pty:", err)
        return
    }
    defer ptyFile.Close()

    go func() {
        buf := make([]byte, 2048)
        for {
            n, err := ptyFile.Read(buf)
            if err != nil {
                return
            }
            ws.WriteMessage(websocket.TextMessage, buf[:n])
        }
    }()

    for {
        _, msg, err := ws.ReadMessage()
        if err != nil {
            return
        }
        ptyFile.Write(msg)
    }
}

func main() {
    port := flag.String("port", "9090", "Listen port")
    flag.Parse()

    http.HandleFunc("/auth", authHandler) 
    http.HandleFunc("/", handleWS)   

    log.Println("Servidor iniciado na porta:", *port)
    http.ListenAndServe("0.0.0.0:"+*port, nil)
}
