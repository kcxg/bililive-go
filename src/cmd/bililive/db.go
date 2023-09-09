package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net"
	"strings"
	"time"
)

var db *sql.DB

type License struct {
	id      int
	user    string
	license string
	mac     string
	expire  sql.NullTime `db:"due_date"`
}

func initDB() (err error) {
	dsn := "current_dev:VLNBHlYTESsS#0X@tcp(43.138.6.131:23399)/license?timeout=2s&readTimeout=6s&interpolateParams=true&parseTime=true"
	db, err = sql.Open("mysql", dsn) // 打开一个数据库连接，不会校验用户名和密码是否正确
	if err != nil {
		return err
	}
	err = db.Ping()

	if err != nil {
		return err
	}
	return
}

// 连接mysql示例
func dbCheck(license string) error {
	//var license string = "1234567"
	mac := GetMac()
	//fmt.Printf("查询结果：%#v\n", mac)
	err := initDB() // 连接数据库
	if err != nil {
		log.Println("连接证书服务器失败!", err)
		return err
	}
	sqlStr := "select id, user, license ,mac, expire from info where license = ?" // 构建查询语句
	var u License
	err = db.QueryRow(sqlStr, license).Scan(&u.id, &u.user, &u.license, &u.mac, &u.expire) // 查询单条数据并扫描结果
	if err != nil {
		if err == sql.ErrNoRows {
			//log.Print("license无效，请联系运营人员获取！")
			return fmt.Errorf("证书失效，请联系运营人员：18611519772（微信），谢谢!\n")
		}
		//log.Printf("查询失败: %s", err)
		return err
	}

	macs := strings.Join(mac, ",")
	if u.mac == "" {
		err := update(macs, license)
		if err != nil {
			return err
		}
	} else {
		macSplit := strings.Split(u.mac, ",")

		for _, m := range mac {
			if In(m, macSplit) == true {
				expire := u.expire.Time.Format("2006-01-02")
				timeNow := time.Now().Format("2006-01-02")
				log.Printf("证书过期时间：%#v\n", expire)
				if timeNow > expire {
					return fmt.Errorf("证书已过期，请联系运营人员: 18611519772（微信） 进行充值!\n")
				} else {
					//fmt.Printf("Licese过期时间：%#v\n", expire)
					return nil
				}
			}
		}
		return fmt.Errorf("证书失效，请联系运营人员：18611519772（微信），谢谢！\n")
	}

	return nil
}

func update(mac string, license string) error {
	sql := "update info set mac = '%s'  where license = '%s'"
	sqlStr := fmt.Sprintf(sql, mac, license)
	//println(sqlStr)
	ret, err := db.Exec(sqlStr)
	if err != nil {
		fmt.Println("更新数据失败", err)
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		fmt.Println("获取行数失败", err)
		return err
	}
	//fmt.Println("更新数据行数：", rows)
	return nil
}

func GetMac() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v\n", err)
		return macAddrs
	}
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}
		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

func In(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}
