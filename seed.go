package wallet


var (
	SeedLong = 15
	SaveSeedLong = 12

	WalletSeed = []byte("walletseed")
	seedlog    = log.New("module", "wallet")

	ChineseSeedCache = make(map[string]string)
	EnglishSeedCache = make(map[string]string)
)

const BACKUPKEYINDEX = "backupkeyindex"

func CreateSeed(folderpath string, lang int32) (string, error) {
	mnem, err := bipwallet.NewMnemonicString(int(lang), 160)
	if err != nil {
		seedlog.Error("CreateSeed", "NewMnemonicString err", err)
		return "", err
	}
	return mnem, nil
}

func InitSeedLibrary() {
	englieshstrs := strings.Split(englishText, " ")
	chinesestrs := strings.Split(chineseText, " ")

	for _, wordstr := range chinesestrs {
		ChineseSeedCache[wordstr] = wordstr
	}

	for _, wordstr := range englieshstrs {
		EnglishSeedCache[wordstr] = wordstr
	}
}

func VerifySeed(seed string) (bool, error) {

	_, err := bipwallet.NewWalletFromMnemonic(bipwallet.TypeCalorie, seed)
	if err != nil {
		seedlog.Error("VerifySeed NewWalletFromMnemonic", "err", err)
		return false, err
	}
	return true, nil
}

func SaveSeedInBatch(db dbm.DB, seed string, password string, batch dbm.Batch) (bool, error) {
	if len(seed) == 0 || len(password) == 0 {
		return false, types.ErrInvalidParam
	}

	Encrypted, err := AesgcmEncrypter([]byte(password), []byte(seed))
	if err != nil {
		seedlog.Error("SaveSeed", "AesgcmEncrypter err", err)
		return false, err
	}
	batch.Set(WalletSeed, Encrypted)
	return true, nil
}

func GetSeed(db dbm.DB, password string) (string, error) {
	if len(password) == 0 {
		return "", types.ErrInvalidParam
	}
	Encryptedseed, err := db.Get(WalletSeed)
	if err != nil {
		return "", err
	}
	if len(Encryptedseed) == 0 {
		return "", types.ErrSeedNotExist
	}
	seed, err := AesgcmDecrypter([]byte(password), Encryptedseed)
	if err != nil {
		seedlog.Error("GetSeed", "AesgcmDecrypter err", err)
		return "", types.ErrInputPassword
	}
	return string(seed), nil
}

func GetPrivkeyBySeed(db dbm.DB, seed string, specificIndex uint32, signType int) (string, error) {
	var backupindex uint32
	var Hexsubprivkey string
	var err error
	var index uint32

	if specificIndex == 0 {
		backuppubkeyindex, err := db.Get([]byte(BACKUPKEYINDEX))
		if backuppubkeyindex == nil || err != nil {
			index = 0
		} else {
			if err = json.Unmarshal(backuppubkeyindex, &backupindex); err != nil {
				return "", err
			}
			index = backupindex + 1
		}
	} else {
		index = specificIndex
	}
	if signType != 1 && signType != 2 {
		return "", types.ErrNotSupport
	}
	if signType == 1 {

		wallet, err := bipwallet.NewWalletFromMnemonic(bipwallet.TypeCalorie, seed)
		if err != nil {
			seedlog.Error("GetPrivkeyBySeed NewWalletFromMnemonic", "err", err)
			wallet, err = bipwallet.NewWalletFromSeed(bipwallet.TypeCalorie, []byte(seed))
			if err != nil {
				seedlog.Error("GetPrivkeyBySeed NewWalletFromSeed", "err", err)
				return "", types.ErrNewWalletFromSeed
			}
		}

		priv, pub, err := wallet.NewKeyPair(index)
		if err != nil {
			seedlog.Error("GetPrivkeyBySeed NewKeyPair", "err", err)
			return "", types.ErrNewKeyPair
		}

		Hexsubprivkey = hex.EncodeToString(priv)

		public, err := bipwallet.PrivkeyToPub(bipwallet.TypeCalorie, priv)
		if err != nil {
			seedlog.Error("GetPrivkeyBySeed PrivkeyToPub", "err", err)
			return "", types.ErrPrivkeyToPub
		}
		if !bytes.Equal(pub, public) {
			seedlog.Error("GetPrivkeyBySeed NewKeyPair pub  != PrivkeyToPub", "err", err)
			return "", types.ErrSubPubKeyVerifyFail
		}

	} else if signType == 2 { 


		var Seed modules.Seed
		hash := common.Sha256([]byte(seed))

		copy(Seed[:], hash)
		sk, _ := sccrypto.GenerateKeyPairDeterministic(sccrypto.HashAll(Seed, index))
		secretKey := fmt.Sprintf("%x", sk)

		Hexsubprivkey = secretKey
	}
	if specificIndex == 0 {
		var pubkeyindex []byte
		pubkeyindex, err = json.Marshal(index)
		if err != nil {
			seedlog.Error("GetPrivkeyBySeed", "Marshal err ", err)
			return "", types.ErrMarshal
		}

		err = db.SetSync([]byte(BACKUPKEYINDEX), pubkeyindex)
		if err != nil {
			seedlog.Error("GetPrivkeyBySeed", "SetSync err ", err)
			return "", err
		}
	}
	return Hexsubprivkey, nil
}

func AesgcmEncrypter(password []byte, seed []byte) ([]byte, error) {
	key := make([]byte, 32)
	if len(password) > 32 {
		key = password[0:32]
	} else {
		copy(key, password)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		seedlog.Error("AesgcmEncrypter NewCipher err", "err", err)
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		seedlog.Error("AesgcmEncrypter NewGCM err", "err", err)
		return nil, err
	}

	Encrypted := aesgcm.Seal(nil, key[:12], seed, nil)
	return Encrypted, nil
}

func AesgcmDecrypter(password []byte, seed []byte) ([]byte, error) {
	key := make([]byte, 32)
	if len(password) > 32 {
		key = password[0:32]
	} else {
		copy(key, password)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		seedlog.Error("AesgcmDecrypter", "NewCipher err", err)
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		seedlog.Error("AesgcmDecrypter", "NewGCM err", err)
		return nil, err
	}
	decryptered, err := aesgcm.Open(nil, key[:12], seed, nil)
	if err != nil {
		seedlog.Error("AesgcmDecrypter", "aesgcm Open err", err)
		return nil, err
	}
	return decryptered, nil
}
