package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/go-pay/gopay/pkg/xlog"
	"net/http"
	"payTest/AliPay/config"
	"time"
)

func main() {
	// 1. 初始化，appid：应用ID；privateKey：应用私钥；isProd：是否是生产环境
	client, err := alipay.NewClient(config.AppId, config.AppPrivateKey, false)
	if err != nil {
		xlog.Error(err)
		return
	}
	client.SetLocation(alipay.LocationShanghai). // 设置时区，不设置或出错均为默认服务器时间
		SetCharset(alipay.UTF8). // 设置字符编码，不设置时，默认 utf-8
		SetSignType(alipay.RSA2). // 设置签名类型，不设置时，默认 RSA2
		SetReturnUrl(config.ReturnURL). // 设置返回URL
		SetNotifyUrl(config.NotifyURL) // 设置异步通知URL
	// 2. 填写交易内容，下面4个set里面的内容是必填的
	outTradeNo := time.Now().Unix()
	xlog.Debugf("outTradeNo: %d\n", outTradeNo)
	bm := make(gopay.BodyMap)
	bm.Set("subject", "测试扫码支付")
	bm.Set("out_trade_no", outTradeNo)
	bm.Set("total_amount", "888")
	bm.Set("product_code", config.ProductCode)
	// 3. 发起交易
	payPageURL, err := client.TradePagePay(context.Background(), bm) // payPageURL是支付宝返回的支付页面
	if err != nil {
		xlog.Error(err)
		return
	}
	xlog.Debugf("payPageURL:%s\n", payPageURL)

	router := gin.Default()
	// 4. 异步接收支付结果
	router.POST("/pay/alipay/notify", func(c *gin.Context) {
		tradeStatus := c.PostForm("trade_status")
		if tradeStatus == "TRADE_CLOSED" {
			c.JSON(http.StatusOK, gin.H{
				"msg": "交易已关闭",
			})
		} else if tradeStatus == "TRADE_SUCCESS" {
			c.JSON(http.StatusOK, gin.H{
				"msg": "交易成功",
			})
		}
	})
	// 5. 同步接收支付结果
	router.GET("/pay/alipay/return", func(c *gin.Context) {
		returnReq, err := alipay.ParseNotifyToBodyMap(c.Request)
		if err != nil {
			xlog.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "参数错误",
			})
			return
		}
		ok, err := alipay.VerifySign(config.AliPayPublicKey, returnReq)
		if err != nil {
			xlog.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "参数错误",
			})
			return
		}
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"msg": "验签失败",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg": "验签成功",
			})
		}
	})
	// 5. 查询支付结果
	router.GET("/pay/alipay/query", func(c *gin.Context) {
		outTradeNo := c.Query("out_trade_no")
		bm := make(gopay.BodyMap)
		bm.Set("out_trade_no", outTradeNo)
		aliRsp, err := client.TradeQuery(context.Background(), bm)
		if err != nil {
			xlog.Error(err)
			c.JSON(http.StatusOK, gin.H{
				"msg": "参数错误",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"msg":  "验签成功",
				"data": aliRsp.Response,
			})
		}

	})

	router.Run()
}
