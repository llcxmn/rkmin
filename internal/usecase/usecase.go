package usecase

import (
	"errors"
	"fmt"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"

	"rkmin/internal/domain"
	"rkmin/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type UploadFunc func(files []*multipart.FileHeader) ([]string, error)

type Usecase struct {
	repo   *repository.Repositories
	upload UploadFunc
}

func New(repo *repository.Repositories, upload UploadFunc) *Usecase {
	return &Usecase{repo: repo, upload: upload}
}

func (u *Usecase) Register(req domain.RegisterRequest) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.KataSandi), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	user := domain.User{
		Nama:         req.Nama,
		KataSandi:    string(hash),
		NoTelp:       req.NoTelp,
		TanggalLahir: req.TanggalLahir,
		Tentang:      req.Tentang,
		Pekerjaan:    req.Pekerjaan,
		Email:        req.Email,
		IDProvinsi:   req.IDProvinsi,
		IDKota:       req.IDKota,
	}
	if err := u.repo.CreateUserWithStore(&user); err != nil {
		return domain.User{}, err
	}
	return u.repo.FindUserByID(user.ID)
}

func (u *Usecase) Login(phone, password string) (domain.User, error) {
	user, err := u.repo.FindUserByPhone(phone)
	if err != nil {
		return domain.User{}, ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(user.KataSandi), []byte(password)) != nil {
		return domain.User{}, ErrUnauthorized
	}
	return user, nil
}

func (u *Usecase) GetUser(userID uint) (domain.User, error) {
	return u.repo.FindUserByID(userID)
}

func (u *Usecase) UpdateUser(userID uint, req domain.UpdateUserRequest) (domain.User, error) {
	user, err := u.repo.FindUserByID(userID)
	if err != nil {
		return domain.User{}, err
	}
	user.Nama = req.Nama
	user.NoTelp = req.NoTelp
	user.TanggalLahir = req.TanggalLahir
	user.Tentang = req.Tentang
	user.Pekerjaan = req.Pekerjaan
	user.Email = req.Email
	user.IDProvinsi = req.IDProvinsi
	user.IDKota = req.IDKota
	if req.KataSandi != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.KataSandi), bcrypt.DefaultCost)
		if err != nil {
			return domain.User{}, err
		}
		user.KataSandi = string(hash)
	}
	if err := u.repo.UpdateUser(&user); err != nil {
		return domain.User{}, err
	}
	return u.repo.FindUserByID(userID)
}

func (u *Usecase) ListAddresses(userID uint, filter repository.AddressFilter) ([]domain.Alamat, error) {
	return u.repo.ListAddresses(userID, filter)
}

func (u *Usecase) GetAddress(userID, id uint) (domain.Alamat, error) {
	return u.repo.FindAddress(userID, id)
}

func (u *Usecase) CreateAddress(userID uint, req domain.AddressRequest) (uint, error) {
	addr := domain.Alamat{
		UserID:       userID,
		JudulAlamat:  req.JudulAlamat,
		NamaPenerima: req.NamaPenerima,
		NoTelp:       req.NoTelp,
		DetailAlamat: req.DetailAlamat,
	}
	if addr.JudulAlamat == "" {
		addr.JudulAlamat = addr.NamaPenerima
	}
	if err := u.repo.CreateAddress(&addr); err != nil {
		return 0, err
	}
	return addr.ID, nil
}

func (u *Usecase) UpdateAddress(userID, id uint, req domain.AddressRequest) error {
	addr, err := u.repo.FindAddress(userID, id)
	if err != nil {
		return err
	}
	if req.JudulAlamat != "" {
		addr.JudulAlamat = req.JudulAlamat
	}
	addr.NamaPenerima = req.NamaPenerima
	addr.NoTelp = req.NoTelp
	addr.DetailAlamat = req.DetailAlamat
	return u.repo.UpdateAddress(&addr)
}

func (u *Usecase) DeleteAddress(userID, id uint) error {
	return u.repo.DeleteAddress(userID, id)
}

func (u *Usecase) ListCategories(filter repository.CategoryFilter) ([]domain.Category, error) {
	return u.repo.ListCategories(filter)
}

func (u *Usecase) GetCategory(id uint) (domain.Category, error) {
	return u.repo.FindCategory(id)
}

func (u *Usecase) CreateCategory(isAdmin bool, req domain.CategoryRequest) (uint, error) {
	if !isAdmin {
		return 0, ErrUnauthorized
	}
	cat := domain.Category{NamaCategory: req.NamaCategory}
	if err := u.repo.CreateCategory(&cat); err != nil {
		return 0, err
	}
	return cat.ID, nil
}

func (u *Usecase) UpdateCategory(isAdmin bool, id uint, req domain.CategoryRequest) error {
	if !isAdmin {
		return ErrUnauthorized
	}
	cat, err := u.repo.FindCategory(id)
	if err != nil {
		return err
	}
	cat.NamaCategory = req.NamaCategory
	return u.repo.UpdateCategory(&cat)
}

func (u *Usecase) DeleteCategory(isAdmin bool, id uint) error {
	if !isAdmin {
		return ErrUnauthorized
	}
	return u.repo.DeleteCategory(id)
}

func (u *Usecase) MyStore(userID uint) (domain.Toko, error) {
	return u.repo.FindMyStore(userID)
}

func (u *Usecase) GetStore(id uint) (domain.Toko, error) {
	return u.repo.FindStore(id)
}

func (u *Usecase) ListStores(name string, page repository.PageFilter) ([]domain.Toko, error) {
	return u.repo.ListStores(name, page)
}

func (u *Usecase) UpdateStore(userID, storeID uint, name string, photoFiles []*multipart.FileHeader) error {
	store, err := u.repo.FindStore(storeID)
	if err != nil {
		return err
	}
	if store.UserID != userID {
		return ErrForbidden
	}
	if name != "" {
		store.NamaToko = name
	}
	if len(photoFiles) > 0 {
		paths, err := u.upload(photoFiles[:1])
		if err != nil {
			return err
		}
		store.URLFoto = paths[0]
	}
	return u.repo.UpdateStore(&store)
}

type ProductInput struct {
	NamaProduk       string
	CategoryID       uint
	HasCategoryID    bool
	HargaReseller    int64
	HasHargaReseller bool
	HargaKonsumen    int64
	HasHargaKonsumen bool
	Stok             int
	HasStok          bool
	Deskripsi        string
	HasDeskripsi     bool
	Photos           []*multipart.FileHeader
}

func (u *Usecase) ListProducts(filter repository.ProductFilter) ([]domain.Product, error) {
	return u.repo.ListProducts(filter)
}

func (u *Usecase) GetProduct(id uint) (domain.Product, error) {
	return u.repo.FindProduct(id)
}

func (u *Usecase) CreateProduct(userID uint, in ProductInput) (uint, error) {
	if strings.TrimSpace(in.NamaProduk) == "" {
		return 0, errors.New("nama_produk is required")
	}
	if !in.HasCategoryID || !in.HasHargaReseller || !in.HasHargaKonsumen || !in.HasStok {
		return 0, errors.New("category_id, harga_reseller, harga_konsumen, and stok are required")
	}
	store, err := u.repo.FindMyStore(userID)
	if err != nil {
		return 0, err
	}
	if _, err := u.repo.FindCategory(in.CategoryID); err != nil {
		return 0, err
	}
	paths, err := u.upload(in.Photos)
	if err != nil {
		return 0, err
	}
	product := domain.Product{
		TokoID:        store.ID,
		CategoryID:    in.CategoryID,
		NamaProduk:    in.NamaProduk,
		Slug:          slugify(fmt.Sprintf("%s-%d", in.NamaProduk, time.Now().UnixNano())),
		HargaReseller: in.HargaReseller,
		HargaKonsumen: in.HargaKonsumen,
		Stok:          in.Stok,
		Deskripsi:     in.Deskripsi,
	}
	for _, path := range paths {
		product.Photos = append(product.Photos, domain.ProductPhoto{URL: path})
	}
	if err := u.repo.CreateProduct(&product); err != nil {
		return 0, err
	}
	if err := u.repo.UpdateProductSlug(product.ID, slugify(fmt.Sprintf("%s-%d", in.NamaProduk, product.ID))); err != nil {
		return 0, err
	}
	return product.ID, nil
}

func (u *Usecase) UpdateProduct(userID, productID uint, in ProductInput) error {
	store, err := u.repo.FindMyStore(userID)
	if err != nil {
		return err
	}
	product, err := u.repo.FindProduct(productID)
	if err != nil {
		return err
	}
	if product.TokoID != store.ID {
		return ErrForbidden
	}
	if in.HasCategoryID {
		if _, err := u.repo.FindCategory(in.CategoryID); err != nil {
			return err
		}
		product.CategoryID = in.CategoryID
	}
	if strings.TrimSpace(in.NamaProduk) != "" {
		product.NamaProduk = in.NamaProduk
		product.Slug = slugify(fmt.Sprintf("%s-%d", in.NamaProduk, product.ID))
	}
	if in.HasHargaReseller {
		product.HargaReseller = in.HargaReseller
	}
	if in.HasHargaKonsumen {
		product.HargaKonsumen = in.HargaKonsumen
	}
	if in.HasStok {
		product.Stok = in.Stok
	}
	if in.HasDeskripsi {
		product.Deskripsi = in.Deskripsi
	}
	replace := len(in.Photos) > 0
	if replace {
		paths, err := u.upload(in.Photos)
		if err != nil {
			return err
		}
		product.Photos = nil
		for _, path := range paths {
			product.Photos = append(product.Photos, domain.ProductPhoto{ProductID: product.ID, URL: path})
		}
	}
	return u.repo.UpdateProduct(&product, replace)
}

func (u *Usecase) DeleteProduct(userID, productID uint) error {
	store, err := u.repo.FindMyStore(userID)
	if err != nil {
		return err
	}
	return u.repo.DeleteProduct(store.ID, productID)
}

func (u *Usecase) ListTransactions(userID uint, filter repository.TransactionFilter) ([]domain.Transaction, error) {
	return u.repo.ListTransactions(userID, filter)
}

func (u *Usecase) GetTransaction(userID, id uint) (domain.Transaction, error) {
	return u.repo.FindTransaction(userID, id)
}

func (u *Usecase) CreateTransaction(userID uint, req domain.TransactionRequest) (uint, error) {
	return u.repo.CreateTransaction(userID, req)
}

func ParseUint(raw string) (uint, error) {
	v, err := strconv.ParseUint(raw, 10, 64)
	return uint(v), err
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = strings.Trim(re.ReplaceAllString(s, "-"), "-")
	if s == "" {
		return "produk"
	}
	return s
}
