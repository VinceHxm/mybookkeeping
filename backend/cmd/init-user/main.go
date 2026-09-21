package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/database"
	"mybookkeeping/internal/service"
)

func main() {
	username := flag.String("username", "", "用户名")
	password := flag.String("password", "", "密码")
	flag.Parse()
	if *username == "" || *password == "" {
		fmt.Println("用法: init-user -username=admin -password=secret")
		os.Exit(1)
	}

	cfg := config.Load()
	db, err := database.OpenMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	auth := service.NewAuthService(db, nil, 0, 2, true)
	user, err := auth.UpsertUserForCLI(*username, *password)
	if err != nil {
		log.Fatalf("创建/重置失败: %v", err)
	}
	cat := service.NewCategoryService(db)
	_ = cat.EnsureDefaults(user.ID)
	tag := service.NewTagService(db)
	_ = tag.EnsureDefaults(user.ID)
	acc := service.NewAccountService(db)
	list, _ := acc.List(user.ID, true)
	if len(list) == 0 {
		_, _ = acc.Create(user.ID, service.AccountInput{Name: "现金", Type: "cash", Sort: 1})
		_, _ = acc.Create(user.ID, service.AccountInput{Name: "银行卡", Type: "bank", Sort: 2})
	}
	fmt.Printf("用户已就绪: id=%d username=%s role=admin（密码已写入/重置，可用该密码登录）\n", user.ID, user.Username)
}
