package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "fmt"
import "log"

func main() {
  fmt.Println("-- Programa de Criar Tabela --")

  fmt.Println("Abrindo conexao...")
  // abrir a conexão
  db, erroAbertura := sql.Open("mysql", "usuario:senha@tcp(localhost:3306)/bancodedados")

  if erroAbertura != nil {
    log.Fatal(erroAbertura.Error())
  }

  erroPing := db.Ping()

  if erroPing != nil {
    log.Fatal(erroPing.Error())
  }

  fmt.Println("Criando tabela...")
  // criar tabela
  _, erroCreate := db.Exec(
    "CREATE TABLE livros (" +
      "id INT NOT NULL AUTO_INCREMENT," +
      "autor VARCHAR(50) NOT NULL," +
      "titulo VARCHAR(100) NOT NULL," +
      "PRIMARY KEY(id)" +
    ")")

  if erroCreate != nil {
    log.Fatal(erroCreate.Error())
  }

  fmt.Println("Criando registros...")
  // criar registros
  var erroInsercao error
  _, erroInsercao = db.Exec(
    "INSERT INTO livros (autor, titulo) VALUES " +
      "('José de Alencar', 'O Guarani')," +
      "('Viriato Correia', 'Cazuza')," +
      "('Machado de Assis', 'Dom Casmurro')")

  if erroInsercao != nil {
    log.Fatal(erroInsercao.Error())
  }

  fmt.Println("Pronto! :)")
}
