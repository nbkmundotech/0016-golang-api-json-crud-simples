package main

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "encoding/json"
  "fmt"
  "github.com/gorilla/mux"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strconv"
)

type Livro struct {
  Id int `json:"id"`
  Titulo string `json:"titulo"`
  Autor string `json:"autor"`
}

type RespostaComErro struct {
  Erro string `json:"erro"`
}

var db *sql.DB

func rotaPrincipal(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Bem vindo")
}

func listarLivros(w http.ResponseWriter, r *http.Request) {
  // Fazer consulta ao banco dados
  registros, erroSelect := db.Query("SELECT id, autor, titulo FROM livros")

  if erroSelect != nil {
    log.Println("listarLivros: erroSelect: " + erroSelect.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  var livros []Livro = make([]Livro, 0)
  for registros.Next() {
    var livro Livro
    erroDeScan := registros.Scan(&livro.Id, &livro.Autor, &livro.Titulo)
    if erroDeScan != nil {
      log.Println("listarLivros: erroDeScan: " + erroDeScan.Error())
      continue
    }

    livros = append(livros, livro)
  }

  erroFecharRegistros := registros.Close()

  if erroFecharRegistros != nil {
    log.Println("listarLivros: erroFecharRegistros: " + erroFecharRegistros.Error())
  }

  encoder := json.NewEncoder(w)
  encoder.Encode(livros)
}

func validarLivro(livro Livro) string {
  if len(livro.Autor) == 0 || len(livro.Autor) > 50 {
    return "Autor tem que ter pelo menos um caractere e no maximo 50 caracteres"
  }
  if len(livro.Titulo) == 0 {
    return "Titulo nao foi fornecido"
  }
  if len(livro.Titulo) > 100 {
    return "Titulo nao pode ter mais de 100 caracteres"
  }

  // Nao houve erro de validacao
  return ""
}

func cadastrarLivro(w http.ResponseWriter, r *http.Request) {
  body, erro := ioutil.ReadAll(r.Body)
  if erro != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  var novoLivro Livro
  json.Unmarshal(body, &novoLivro)

  // Validacao
  erroValidacao := validarLivro(novoLivro)
  if len(erroValidacao) > 0 {
    w.WriteHeader(http.StatusUnprocessableEntity)
    json.NewEncoder(w).Encode(RespostaComErro{erroValidacao})
    return
  }

  // Insercao no Banco de Dados
  resultado, erroDeInsert := db.Exec("INSERT INTO livros (autor, titulo) VALUES (?, ?)", novoLivro.Autor, novoLivro.Titulo)

  idGerado, erroLastInsertId := resultado.LastInsertId()

  if erroDeInsert != nil || erroLastInsertId != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  novoLivro.Id = int(idGerado)

  w.WriteHeader(http.StatusCreated)
  encoder := json.NewEncoder(w)
  encoder.Encode(novoLivro)
}

func excluirLivro(w http.ResponseWriter, r *http.Request) {
  // e.g. DELETE /livros/123
  vars := mux.Vars(r)
  id, erro := strconv.Atoi(vars["livroId"])

  if erro != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  registro := db.QueryRow("SELECT id FROM livros WHERE id = ?", id)
  var idDoLivro int
  erroDeScan := registro.Scan(&idDoLivro)

  if erroDeScan != nil {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  _, erroDeExec := db.Exec("DELETE FROM livros WHERE id = ?", id)

  if erroDeExec != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusNoContent)
}

func modificarLivro(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, erro := strconv.Atoi(vars["livroId"])

  if erro != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  corpo, erroCorpo := ioutil.ReadAll(r.Body)

  if erroCorpo != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  var livroModificado Livro
  erroJson := json.Unmarshal(corpo, &livroModificado)

  if erroJson != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  registro := db.QueryRow("SELECT id, autor, titulo FROM livros WHERE id = ?", id)
  var livro Livro
  erroScan := registro.Scan(&livro.Id, &livro.Autor, &livro.Titulo)

  if erroScan != nil {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  _, erroExec := db.Exec("UPDATE livros SET autor = ?, titulo = ? WHERE id = ?", livroModificado.Autor, livroModificado.Titulo, id)

  if erroExec != nil {
    log.Println("modificarLivro: erroExec: " + erroExec.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  json.NewEncoder(w).Encode(livroModificado)
}

func buscarLivro(w http.ResponseWriter, r *http.Request) {
  // "/livros/123" --> ["", "livros", "123"]

  vars := mux.Vars(r)
  id, _ := strconv.Atoi(vars["livroId"])
  // exercicio: lidar com erro

  registro := db.QueryRow("SELECT id, autor, titulo FROM livros WHERE id = ?", id)
  var livro Livro
  erroScan := registro.Scan(&livro.Id, &livro.Autor, &livro.Titulo)

  if erroScan != nil {
    log.Println("buscarLivro: erroScan: " + erroScan.Error())
    w.WriteHeader(http.StatusNotFound)
    return
  }

  json.NewEncoder(w).Encode(livro)
}

func configurarRotas(roteador *mux.Router) {
  roteador.HandleFunc("/", rotaPrincipal)
  roteador.HandleFunc("/livros", listarLivros).Methods("GET")
  // e.g. GET /livros/123
  roteador.HandleFunc("/livros/{livroId}", buscarLivro).Methods("GET")
  roteador.HandleFunc("/livros", cadastrarLivro).Methods("POST")
  roteador.HandleFunc("/livros/{livroId}", modificarLivro).Methods("PUT")
  roteador.HandleFunc("/livros/{livroId}", excluirLivro).Methods("DELETE")
}

func jsonMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    next.ServeHTTP(w, r)
  })
}

func configurarServidor() {
  roteador := mux.NewRouter().StrictSlash(true)
  roteador.Use(jsonMiddleware)
  configurarRotas(roteador)

  fmt.Println("Servidor esta rodando na porta 1337")
  log.Fatal(http.ListenAndServe(":1337", roteador))
}

func configurarBancoDeDados() {
  var erroAbertura error
  var stringDeConexao string = fmt.Sprintf(
    "%s:%s@tcp(%s)/%s",
    os.Getenv("DB_USUARIO"),
    os.Getenv("DB_SENHA"),
    os.Getenv("DB_HOST_COM_PORTA"),
    os.Getenv("DB_BANCO_DE_DADOS"),
  )
  db, erroAbertura = sql.Open("mysql", stringDeConexao)

  if erroAbertura != nil {
    log.Fatal(erroAbertura.Error())
  }

  erroPing := db.Ping()

  if erroPing != nil {
    log.Fatal(erroPing.Error())
  }
}

func main() {
  configurarBancoDeDados()
  configurarServidor()
}
