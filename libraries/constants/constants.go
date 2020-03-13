package constants

const (
	AES256_PASSPHRASE string = "0123456789abcdef"
	HASH_TOKEN        string = "h5%@cala%games"
)

const (
	FIRE    string = "f"
	EARTH   string = "e"
	THUNDER string = "t"
	WATER   string = "w"
	LIGHT   string = "l"
	DARK    string = "d"
	EVOLVE  string = "evo" //Tiến hóa
)

const (
	ID       string = "id"
	TIME     string = "time"
	LIMIT    string = "limit"
	COST     string = "cost"
	PRICE    string = "price"
	BOUGHT   string = "bought"
	INDEX    string = "index"
	GIFT     string = "gift"
	PERCENT  string = "percent"
	RANDOM   string = "random"
	RAND     string = "rand"
	ALL      string = "all"
	LEVEL    string = "level"
	BACk     string = "back"
	QUANTITY string = "quantity"
	TYPE     string = "type"
	BRANCH   string = "branch"

	STONES              string = "stones"
	EXP                 string = "exp"
	GOLD                string = "gold"
	GEM                 string = "gem"
	PIECE               string = "piece"
	PIECE_GENERAL       string = "piece_general"
	EQUIP               string = "equip"
	TICKET_ARENA        string = "ticket_arena"
	POWER_TOWER         string = "power_tower"      //=======(updateGifts)
	TICKET_ARENA_PVP    string = "ticket_arena_pvp" //=======(updateGifts)
	SUMMON_BALL         string = "summon_ball"      //1:Lục 2:Tím 3:Đỏ
	TICKET_MARKET_BLACK string = "ticket_market_black"
	TICKET_GUARDIAN     string = "ticket_guardian" //Ngọc Lục Bảo
	FLOWER_GUARDIAN     string = "flower_guardian"
	FRUIT_GUARDIAN      string = "fruit_guardian"
	STONE_GUARDIAN      string = "stone_guardian"
	PIECE_GUARDIAN      string = "piece_guardian"
	AVATAR              string = "avatar" // cân nhắc đổi chuẩn nhận avatar như pokiwar => tạm bỏ rồi chuyển qua pet_id
	GEM_SOUL            string = "gem_soul"
	ARENA_COIN          string = "arena_coin"
)

const (
	INBOX_SENDER_BY_SYSTEM  uint8 = 1
	INBOX_SENDER_BY_CLUB    uint8 = 2
	INBOX_SENDER_BY_EVENT   uint8 = 3
	INBOX_SENDER_BY_MESSAGE uint8 = 4
)

const (
	PAYMENT_INGAME uint8 = 1
	PAYMENT_INWEB  uint8 = 2

	ASC  int = 1
	DESC int = -1
)
