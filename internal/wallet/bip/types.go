package bip

// DerivationPath BIP44 派生路径
// 格式: m/purpose'/coin_type'/account'/change/address_index
// 例如: m/44'/60'/0'/0/0 (Ethereum 第一个地址)
type DerivationPath struct {
	Purpose     uint32 // 44 (BIP44)
	CoinType    uint32 // 60 (Ethereum), 0 (Bitcoin), 501 (Solana)
	Account     uint32 // 账户索引
	Change      uint32 // 0 (外部链), 1 (找零链)
	AddressIndex uint32 // 地址索引
}

// WalletInfo 钱包信息
type WalletInfo struct {
	Mnemonic    string   // 助记词（BIP39）
	Seed        []byte   // 种子（从助记词派生）
	MasterKey   []byte   // 主私钥（BIP32）
	ChainCode   []byte   // 链码
	Derivations map[string]*KeyPair // 派生路径 -> 密钥对
}

// KeyPair 密钥对
type KeyPair struct {
	PrivateKey []byte // 私钥（不加密，实际使用时需要加密存储）
	PublicKey  []byte // 公钥
	Address    string // 地址
	Path       DerivationPath // 派生路径
}

// AddressInfo 地址信息
type AddressInfo struct {
	Address     string
	PrivateKey  []byte // 加密后的私钥
	PublicKey   []byte
	Path        DerivationPath
	Chain       string
	CreatedAt   int64
}

