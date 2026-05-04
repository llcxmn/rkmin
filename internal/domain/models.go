package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Nama         string         `gorm:"size:255;not null" json:"nama"`
	KataSandi    string         `gorm:"size:255;not null" json:"-"`
	NoTelp       string         `gorm:"size:30;uniqueIndex;not null" json:"no_telp"`
	TanggalLahir string         `gorm:"size:30" json:"tanggal_Lahir"`
	Tentang      string         `gorm:"type:text" json:"tentang"`
	Pekerjaan    string         `gorm:"size:255" json:"pekerjaan"`
	Email        string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	IDProvinsi   string         `gorm:"size:16" json:"id_provinsi"`
	IDKota       string         `gorm:"size:16" json:"id_kota"`
	IsAdmin      bool           `gorm:"default:false" json:"is_admin"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Toko         Toko           `json:"toko,omitempty"`
	Alamat       []Alamat       `json:"alamat,omitempty"`
}

type Toko struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	NamaToko  string         `gorm:"size:255;not null" json:"nama_toko"`
	URLFoto   string         `gorm:"size:500" json:"url_foto"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      *User          `json:"user,omitempty"`
	Products  []Product      `json:"products,omitempty"`
}

type Alamat struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	JudulAlamat  string         `gorm:"size:255;not null" json:"judul_alamat"`
	NamaPenerima string         `gorm:"size:255;not null" json:"nama_penerima"`
	NoTelp       string         `gorm:"size:30;not null" json:"no_telp"`
	DetailAlamat string         `gorm:"type:text;not null" json:"detail_alamat"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Category struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	NamaCategory string         `gorm:"size:255;uniqueIndex;not null" json:"nama_category"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Product struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TokoID        uint           `gorm:"index;not null" json:"toko_id"`
	CategoryID    uint           `gorm:"index;not null" json:"category_id"`
	NamaProduk    string         `gorm:"size:255;not null" json:"nama_produk"`
	Slug          string         `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	HargaReseller int64          `gorm:"not null" json:"harga_reseller"`
	HargaKonsumen int64          `gorm:"not null" json:"harga_konsumen"`
	Stok          int            `gorm:"not null" json:"stok"`
	Deskripsi     string         `gorm:"type:text" json:"deskripsi"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Toko          Toko           `json:"toko"`
	Category      Category       `json:"category"`
	Photos        []ProductPhoto `json:"photos"`
}

type ProductPhoto struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProductID uint           `gorm:"index;not null" json:"product_id"`
	URL       string         `gorm:"size:500;not null" json:"url"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Transaction struct {
	ID          uint                `gorm:"primaryKey" json:"id"`
	UserID      uint                `gorm:"index;not null" json:"user_id"`
	AlamatID    uint                `gorm:"index;not null" json:"alamat_kirim"`
	MethodBayar string              `gorm:"size:100;not null" json:"method_bayar"`
	HargaTotal  int64               `gorm:"not null" json:"harga_total"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `gorm:"index" json:"-"`
	User        User                `json:"user"`
	Alamat      Alamat              `json:"alamat"`
	Details     []TransactionDetail `json:"detail_trx"`
}

type TransactionDetail struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	TransactionID uint       `gorm:"index;not null" json:"transaction_id"`
	TokoID        uint       `gorm:"index;not null" json:"toko_id"`
	ProductLogID  uint       `gorm:"index;not null" json:"product_log_id"`
	Kuantitas     int        `gorm:"not null" json:"kuantitas"`
	HargaTotal    int64      `gorm:"not null" json:"harga_total"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Toko          Toko       `json:"toko"`
	Product       ProductLog `gorm:"foreignKey:ProductLogID" json:"product"`
}

type ProductLog struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ProductID     uint           `gorm:"index;not null" json:"product_id"`
	TokoID        uint           `gorm:"index;not null" json:"toko_id"`
	CategoryID    uint           `gorm:"index;not null" json:"category_id"`
	NamaProduk    string         `gorm:"size:255;not null" json:"nama_produk"`
	Slug          string         `gorm:"size:255;not null" json:"slug"`
	HargaReseller int64          `gorm:"not null" json:"harga_reseller"`
	HargaKonsumen int64          `gorm:"not null" json:"harga_konsumen"`
	Deskripsi     string         `gorm:"type:text" json:"deskripsi"`
	CreatedAt     time.Time      `json:"created_at"`
	Toko          Toko           `json:"toko"`
	Category      Category       `json:"category"`
	PhotosJSON    string         `gorm:"type:json" json:"-"`
	Photos        []ProductPhoto `gorm:"-" json:"photos"`
}

func (p *ProductLog) AfterFind(tx *gorm.DB) error {
	if p.PhotosJSON == "" {
		return nil
	}
	_ = json.Unmarshal([]byte(p.PhotosJSON), &p.Photos)
	return nil
}

func AutoMigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Toko{},
		&Alamat{},
		&Category{},
		&Product{},
		&ProductPhoto{},
		&ProductLog{},
		&Transaction{},
		&TransactionDetail{},
	)
}
