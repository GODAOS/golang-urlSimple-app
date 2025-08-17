# -*- coding: utf-8 -*-

package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 데이터베이스 연결을 위한 전역 변수
var db *sql.DB

// 단축 URL의 길이
const shortURLLength = 6

// 단축 URL 생성 함수
func generateShortURL() string {
	// 랜덤 시드 설정
	rand.Seed(time.Now().UnixNano())
	// 단축 URL에 사용될 문자들
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 단축 URL을 저장할 슬라이스
	b := make([]byte, shortURLLength)
	// 문자셋에서 랜덤하게 문자를 선택하여 슬라이스에 추가
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// 리디렉션 핸들러 함수
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// URL 경로에서 단축 URL 추출
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	// 단축 URL이 없는 경우 메인 페이지 안내
	if shortURL == "" {
		fmt.Fprintf(w, "URL Shortener is running!")
		return
	}

	// 데이터베이스에서 긴 URL 조회
	var longURL string
	err := db.QueryRow("SELECT long_url FROM urls WHERE short_url = ?", shortURL).Scan(&longURL)
	// URL을 찾지 못한 경우 404 에러 반환
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 원래의 긴 URL로 리디렉션
	http.Redirect(w, r, longURL, http.StatusFound)
}

// 메인 함수
func main() {
	var err error
	// SQLite 데이터베이스 열기
	db, err = sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// urls 테이블이 없는 경우 생성
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (short_url TEXT PRIMARY KEY, long_url TEXT)`);
	if err != nil {
		log.Fatal(err)
	}

	// 리디렉션 핸들러 등록
	http.HandleFunc("/", redirectHandler)

	// 단축 핸들러 등록
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		// POST 메서드만 허용
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// 폼에서 긴 URL 가져오기
		longURL := r.FormValue("url")
		if longURL == "" {
			http.Error(w, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		// 단축 URL 생성
		shortURL := generateShortURL()
		// 데이터베이스에 단축 URL과 긴 URL 저장
		_, err := db.Exec("INSERT INTO urls (short_url, long_url) VALUES (?, ?)", shortURL, longURL)
		if err != nil {
			http.Error(w, "Failed to shorten URL", http.StatusInternalServerError)
			return
		}

		// 생성된 단축 URL 반환
		fmt.Fprintf(w, "Short URL: http://localhost:8080/%s", shortURL)
	})

	log.Println("Server starting on port 8080...")
	// 서버 시작
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
