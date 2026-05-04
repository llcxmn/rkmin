package repository

import (
	"errors"
	"fmt"
	"strings"

	"rkmin/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("record not found")

type PageFilter struct {
	Page  int
	Limit int
}

func (p PageFilter) Normalize() PageFilter {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return p
}

func (p PageFilter) Scope(db *gorm.DB) *gorm.DB {
	p = p.Normalize()
	return db.Offset((p.Page - 1) * p.Limit).Limit(p.Limit)
}

type Repositories struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{DB: db}
}

func (r *Repositories) CreateUserWithStore(user *domain.User) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		store := domain.Toko{UserID: user.ID, NamaToko: fmt.Sprintf("Toko %s", user.Nama)}
		if strings.TrimSpace(store.NamaToko) == "Toko" {
			store.NamaToko = fmt.Sprintf("Toko User %d", user.ID)
		}
		return tx.Create(&store).Error
	})
}

func (r *Repositories) FindUserByPhone(phone string) (domain.User, error) {
	var user domain.User
	err := r.DB.Where("no_telp = ?", phone).First(&user).Error
	return user, normalizeErr(err)
}

func (r *Repositories) FindUserByID(id uint) (domain.User, error) {
	var user domain.User
	err := r.DB.Preload("Toko").First(&user, id).Error
	return user, normalizeErr(err)
}

func (r *Repositories) UpdateUser(user *domain.User) error {
	return r.DB.Save(user).Error
}

func (r *Repositories) ListAddresses(userID uint, page PageFilter) ([]domain.Alamat, error) {
	var rows []domain.Alamat
	err := page.Scope(r.DB.Where("user_id = ?", userID).Order("id desc")).Find(&rows).Error
	return rows, err
}

func (r *Repositories) FindAddress(userID, id uint) (domain.Alamat, error) {
	var row domain.Alamat
	err := r.DB.Where("user_id = ? AND id = ?", userID, id).First(&row).Error
	return row, normalizeErr(err)
}

func (r *Repositories) CreateAddress(addr *domain.Alamat) error {
	return r.DB.Create(addr).Error
}

func (r *Repositories) UpdateAddress(addr *domain.Alamat) error {
	return r.DB.Save(addr).Error
}

func (r *Repositories) DeleteAddress(userID, id uint) error {
	res := r.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&domain.Alamat{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repositories) ListCategories() ([]domain.Category, error) {
	var rows []domain.Category
	return rows, r.DB.Order("id asc").Find(&rows).Error
}

func (r *Repositories) FindCategory(id uint) (domain.Category, error) {
	var row domain.Category
	err := r.DB.First(&row, id).Error
	return row, normalizeErr(err)
}

func (r *Repositories) CreateCategory(cat *domain.Category) error {
	return r.DB.Create(cat).Error
}

func (r *Repositories) UpdateCategory(cat *domain.Category) error {
	return r.DB.Save(cat).Error
}

func (r *Repositories) DeleteCategory(id uint) error {
	res := r.DB.Delete(&domain.Category{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repositories) FindMyStore(userID uint) (domain.Toko, error) {
	var row domain.Toko
	err := r.DB.Where("user_id = ?", userID).First(&row).Error
	return row, normalizeErr(err)
}

func (r *Repositories) FindStore(id uint) (domain.Toko, error) {
	var row domain.Toko
	err := r.DB.Preload("User").First(&row, id).Error
	return row, normalizeErr(err)
}

func (r *Repositories) ListStores(name string, page PageFilter) ([]domain.Toko, error) {
	var rows []domain.Toko
	q := r.DB.Order("id desc")
	if name != "" {
		q = q.Where("nama_toko LIKE ?", "%"+name+"%")
	}
	return rows, page.Scope(q).Find(&rows).Error
}

func (r *Repositories) UpdateStore(store *domain.Toko) error {
	return r.DB.Save(store).Error
}

type ProductFilter struct {
	NamaProduk string
	CategoryID uint
	TokoID     uint
	MinHarga   int64
	MaxHarga   int64
	Page       PageFilter
}

func (r *Repositories) ListProducts(f ProductFilter) ([]domain.Product, error) {
	var rows []domain.Product
	q := r.productPreload().Order("products.id desc")
	if f.NamaProduk != "" {
		q = q.Where("nama_produk LIKE ?", "%"+f.NamaProduk+"%")
	}
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.TokoID > 0 {
		q = q.Where("toko_id = ?", f.TokoID)
	}
	if f.MinHarga > 0 {
		q = q.Where("harga_konsumen >= ?", f.MinHarga)
	}
	if f.MaxHarga > 0 {
		q = q.Where("harga_konsumen <= ?", f.MaxHarga)
	}
	return rows, f.Page.Scope(q).Find(&rows).Error
}

func (r *Repositories) FindProduct(id uint) (domain.Product, error) {
	var row domain.Product
	err := r.productPreload().First(&row, id).Error
	return row, normalizeErr(err)
}

func (r *Repositories) CreateProduct(product *domain.Product) error {
	return r.DB.Create(product).Error
}

func (r *Repositories) UpdateProduct(product *domain.Product, replacePhotos bool) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if replacePhotos {
			if err := tx.Where("product_id = ?", product.ID).Delete(&domain.ProductPhoto{}).Error; err != nil {
				return err
			}
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(product).Error
	})
}

func (r *Repositories) DeleteProduct(storeID, id uint) error {
	res := r.DB.Where("toko_id = ? AND id = ?", storeID, id).Delete(&domain.Product{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repositories) ListTransactions(userID uint, page PageFilter) ([]domain.Transaction, error) {
	var rows []domain.Transaction
	q := r.trxPreload().Where("user_id = ?", userID).Order("id desc")
	return rows, page.Scope(q).Find(&rows).Error
}

func (r *Repositories) FindTransaction(userID, id uint) (domain.Transaction, error) {
	var row domain.Transaction
	err := r.trxPreload().Where("user_id = ? AND id = ?", userID, id).First(&row).Error
	return row, normalizeErr(err)
}

func (r *Repositories) CreateTransaction(userID uint, req domain.TransactionRequest) (uint, error) {
	var trxID uint
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		var address domain.Alamat
		if err := tx.Where("id = ? AND user_id = ?", req.AlamatKirim, userID).First(&address).Error; err != nil {
			return normalizeErr(err)
		}

		trx := domain.Transaction{
			UserID:      userID,
			AlamatID:    req.AlamatKirim,
			MethodBayar: req.MethodBayar,
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}
		var grandTotal int64
		for _, item := range req.DetailTRX {
			var product domain.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Preload("Photos").First(&product, item.ProductID).Error; err != nil {
				return normalizeErr(err)
			}
			if product.Stok < item.Kuantitas {
				return fmt.Errorf("stok produk %s tidak mencukupi", product.NamaProduk)
			}
			product.Stok -= item.Kuantitas
			if err := tx.Model(&domain.Product{}).Where("id = ?", product.ID).Update("stok", product.Stok).Error; err != nil {
				return err
			}
			log := domain.ProductLog{
				ProductID:     product.ID,
				TokoID:        product.TokoID,
				CategoryID:    product.CategoryID,
				NamaProduk:    product.NamaProduk,
				Slug:          product.Slug,
				HargaReseller: product.HargaReseller,
				HargaKonsumen: product.HargaKonsumen,
				Deskripsi:     product.Deskripsi,
				PhotosJSON:    photosJSON(product.Photos),
			}
			if err := tx.Create(&log).Error; err != nil {
				return err
			}
			total := int64(item.Kuantitas) * product.HargaKonsumen
			detail := domain.TransactionDetail{
				TransactionID: trx.ID,
				TokoID:        product.TokoID,
				ProductLogID:  log.ID,
				Kuantitas:     item.Kuantitas,
				HargaTotal:    total,
			}
			if err := tx.Create(&detail).Error; err != nil {
				return err
			}
			grandTotal += total
		}
		if err := tx.Model(&trx).Update("harga_total", grandTotal).Error; err != nil {
			return err
		}
		trxID = trx.ID
		return nil
	})
	return trxID, err
}

func (r *Repositories) productPreload() *gorm.DB {
	return r.DB.Preload("Toko").Preload("Category").Preload("Photos")
}

func (r *Repositories) trxPreload() *gorm.DB {
	return r.DB.Preload("Alamat").
		Preload("Details.Toko").
		Preload("Details.Product.Toko").
		Preload("Details.Product.Category")
}

func normalizeErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func photosJSON(photos []domain.ProductPhoto) string {
	if len(photos) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(photos))
	for _, photo := range photos {
		parts = append(parts, fmt.Sprintf(`{"id":%d,"product_id":%d,"url":%q}`, photo.ID, photo.ProductID, photo.URL))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
