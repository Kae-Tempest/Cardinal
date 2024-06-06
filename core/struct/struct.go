package _struct

type Player struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ServerID string `json:"server_id"`
	Username string `json:"username"`
	RaceID   int    `json:"race_id"`
	JobID    int    `json:"job_id"`
	Exp      int    `json:"exp"`
	Po       int    `json:"po"`
	Level    int    `json:"level"`
	GuildID  int    `json:"guild_id"` // 0 = no guild
}

type Inventory struct {
	UserID   int `json:"user_id"`
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

type Job struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"` // Description of the job
}

type Race struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"` // Description of the Race
}

type Items struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        int    `json:"type"` // 0 = Equipable, 1 = Consomable, 2 = Quest
}

//const (
//	Equipable  int = 0
//	Consomable int = 1
//	Quest      int = 2
//)

type Guild struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Members []int  `json:"members"`
	Owner   string `json:"owner"`
}

type Skill struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	JobID       int    `json:"job_id"`
	AccessAll   bool   `json:"access_all"` // If true, all jobs can use this skill
	Description string `json:"description"`
}

type User_Pet struct {
	PetID  int `json:"pet_id"`
	UserID int `json:"user_id"`
}
type Summon_Beast struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Name         string `json:"name"`
	Strength     int    `json:"strength"`
	Constitution int    `json:"constitution"`
	Mana         int    `json:"mana"`
	Stamina      int    `json:"stamina"`
	Dexterity    int    `json:"dexterity"`
	Intelligence int    `json:"intelligence"`
	Wisdom       int    `json:"wisdom"`
	Charisma     int    `json:"charisma"`
}
type Stats struct {
	UserID       int `json:"user_id"`
	Strength     int `json:"strength"`
	Constitution int `json:"constitution"`
	Mana         int `json:"mana"`
	Stamina      int `json:"stamina"`
	Dexterity    int `json:"dexterity"`
	Intelligence int `json:"intelligence"`
	Charisma     int `json:"charisma"`
}
type Pets struct {
	CreatureID  int  `json:"creature_id"`
	IsMoumtable bool `json:"is_mountable"`
	Speed       int  `json:"speed"` // 0 = slow, 1 = normal, 2 = fast
}
type Effects struct {
	ReferenceID  int `json:"reference_id"` // Item ID or Skill ID or pet ID or creature ID
	Strength     int `json:"strength"`
	Constitution int `json:"constitution"`
	Mana         int `json:"mana"`
	Stamina      int `json:"stamina"`
	Dexterity    int `json:"dexterity"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
	Use          int `json:"use"` // 0 = item, 1 = skill, 2 = pet, 3 = creature
}
type Equipment struct {
	UserID      int `json:"user_id"`
	Helmet      int `json:"helmet"`
	Chestplate  int `json:"chestplate"`
	Leggings    int `json:"leggings"`
	Boots       int `json:"boots"`
	MainHand    int `json:"main_hand"`
	OffHand     int `json:"off_hand"`
	Accesorry_0 int `json:"accesorry_0"`
	Accesorry_1 int `json:"accesorry_1"`
	Accesorry_2 int `json:"accesorry_2"`
	Accesorry_3 int `json:"accesorry_3"`
}

type Creature struct {
	ID           int    `json:"id"`
	Mame         string `json:"name"`
	IsPet        bool   `json:"is_pet"`
	Strength     int    `json:"strength"`
	Constitution int    `json:"constitution"`
	Mana         int    `json:"mana"`
	Stamina      int    `json:"stamina"`
	Dexterity    int    `json:"dexterity"`
	Intelligence int    `json:"intelligence"`
	Charisma     int    `json:"charisma"`
	Wisdom       int    `json:"wisdom"`
}
type Quests struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	IsGroup     bool        `json:"is_group"`
	Difficulty  int         `json:"difficulty"`
	Data        []Objective `json:"data"`
	Reward      Reward      `json:"reward"`
}

type Reward struct {
	Exp  int   `json:"exp"`
	Po   int   `json:"po"`
	Item []int `json:"item"`
}

type Objective struct {
	Title        string `json:"title"`         // {"objectif": "tuer 10 monstres"}
	Objective    int    `json:"objective"`     // {"track": 0}
	MaxObjective int    `json:"max_objective"` // {"max_track": 10}
}
