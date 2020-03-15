package test


func TestBipwallet(t *testing.T) {

	    TypeEther:        "ETH",
		TypeEtherClassic: "ETC",
		TypeBitcoin:      "BTC",
		TypeLitecoin:     "LTC",
		TypeZayedcoin:    "ZEC",
		TypeCalorie:      "CAL",
		TypeYcc:          "YCC",
	*/
	mnem, err := bipwallet.NewMnemonicString(0, 160)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("password:", mnem)
	wallet, err := bipwallet.NewWalletFromMnemonic(bipwallet.TypeEther,
		"wish address cram damp very indicate regret sound figure scheme review scout")
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	var index uint32
	priv, pub, err := wallet.NewKeyPair(index)
	fmt.Println("privkey:", hex.EncodeToString(priv))
	fmt.Println("pubkey:", hex.EncodeToString(pub))
	address, err := wallet.NewAddress(index)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}

	fmt.Println("address:", address)
	address, err = bipwallet.PubToAddress(bipwallet.TypeEther, pub)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}

	fmt.Println("PubToAddress:", address)

	pub, err = bipwallet.PrivkeyToPub(bipwallet.TypeEther, priv)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}

	fmt.Println("PrivToPub:", hex.EncodeToString(pub))

}
