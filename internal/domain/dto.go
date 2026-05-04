package domain

type RegisterRequest struct {
	Nama         string `json:"nama" binding:"required"`
	KataSandi    string `json:"kata_sandi" binding:"required,min=6"`
	NoTelp       string `json:"no_telp" binding:"required"`
	TanggalLahir string `json:"tanggal_Lahir"`
	Tentang      string `json:"tentang"`
	Pekerjaan    string `json:"pekerjaan"`
	Email        string `json:"email" binding:"required,email"`
	IDProvinsi   string `json:"id_provinsi"`
	IDKota       string `json:"id_kota"`
}

type LoginRequest struct {
	NoTelp    string `json:"no_telp" binding:"required"`
	KataSandi string `json:"kata_sandi" binding:"required"`
}

type UpdateUserRequest RegisterRequest

type AddressRequest struct {
	JudulAlamat  string `json:"judul_alamat"`
	NamaPenerima string `json:"nama_penerima" binding:"required"`
	NoTelp       string `json:"no_telp" binding:"required"`
	DetailAlamat string `json:"detail_alamat" binding:"required"`
}

type CategoryRequest struct {
	NamaCategory string `json:"nama_category" binding:"required"`
}

type TransactionRequest struct {
	MethodBayar string                     `json:"method_bayar" binding:"required"`
	AlamatKirim uint                       `json:"alamat_kirim" binding:"required"`
	DetailTRX   []TransactionDetailRequest `json:"detail_trx" binding:"required,min=1"`
}

type TransactionDetailRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Kuantitas int  `json:"kuantitas" binding:"required,min=1"`
}

type AuthResponse struct {
	Nama         string `json:"nama"`
	NoTelp       string `json:"no_telp"`
	TanggalLahir string `json:"tanggal_Lahir"`
	Tentang      string `json:"tentang"`
	Pekerjaan    string `json:"pekerjaan"`
	Email        string `json:"email"`
	IDProvinsi   string `json:"id_provinsi"`
	IDKota       string `json:"id_kota"`
	Token        string `json:"token"`
}
