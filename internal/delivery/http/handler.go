package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"rkmin/internal/domain"
	"rkmin/internal/provcity"
	"rkmin/internal/repository"
	"rkmin/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc       *usecase.Usecase
	jwt      *JWTService
	provcity *provcity.Client
}

func NewHandler(uc *usecase.Usecase, jwt *JWTService, provcity *provcity.Client) *Handler {
	return &Handler{uc: uc, jwt: jwt, provcity: provcity}
}

func (h *Handler) RegisterRoutes(r *gin.Engine, base string) {
	api := r.Group(base)
	api.POST("/auth/register", h.register)
	api.POST("/auth/login", h.login)

	api.GET("/category", h.listCategories)
	api.GET("/category/:id", h.getCategory)

	api.GET("/toko", h.listStores)
	api.GET("/toko/:id_toko", h.getStoreDynamic)

	api.GET("/product", h.listProducts)
	api.GET("/product/:id", h.getProduct)

	api.GET("/provcity/listprovincies", h.listProvinces)
	api.GET("/provcity/listcities/:prov_id", h.listCities)
	api.GET("/provcity/detailprovince/:prov_id", h.detailProvince)
	api.GET("/provcity/detailcity/:city_id", h.detailCity)

	auth := api.Group("", h.jwt.Middleware())
	auth.GET("/user", h.getProfile)
	auth.PUT("/user", h.updateProfile)
	auth.GET("/user/alamat", h.listAddresses)
	auth.GET("/user/alamat/:id", h.getAddress)
	auth.POST("/user/alamat", h.createAddress)
	auth.PUT("/user/alamat/:id", h.updateAddress)
	auth.DELETE("/user/alamat/:id", h.deleteAddress)

	auth.POST("/category", h.createCategory)
	auth.PUT("/category/:id", h.updateCategory)
	auth.DELETE("/category/:id", h.deleteCategory)

	auth.PUT("/toko/:id_toko", h.updateStore)

	auth.POST("/product", h.createProduct)
	auth.PUT("/product/:id", h.updateProduct)
	auth.DELETE("/product/:id", h.deleteProduct)

	auth.GET("/trx", h.listTransactions)
	auth.GET("/trx/:id", h.getTransaction)
	auth.POST("/trx", h.createTransaction)
}

func (h *Handler) register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	user, err := h.uc.Register(req)
	if err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	token, err := h.jwt.Sign(user)
	if err != nil {
		fail(c, http.StatusInternalServerError, "POST", err)
		return
	}
	ok(c, "POST", authResponse(user, token))
}

func (h *Handler) login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	user, err := h.uc.Login(req.NoTelp, req.KataSandi)
	if err != nil {
		fail(c, http.StatusUnauthorized, "POST", errors.New("No Telp atau kata sandi salah"))
		return
	}
	token, err := h.jwt.Sign(user)
	if err != nil {
		fail(c, http.StatusInternalServerError, "POST", err)
		return
	}
	ok(c, "POST", authResponse(user, token))
}

func (h *Handler) getProfile(c *gin.Context) {
	user, err := h.uc.GetUser(authUser(c).ID)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", user)
}

func (h *Handler) updateProfile(c *gin.Context) {
	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	user, err := h.uc.UpdateUser(authUser(c).ID, req)
	if err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	ok(c, "PUT", user)
}

func (h *Handler) listAddresses(c *gin.Context) {
	p := page(c).Normalize()
	filter := repository.AddressFilter{
		Search:       c.Query("search"),
		JudulAlamat:  c.Query("judul_alamat"),
		NamaPenerima: c.Query("nama_penerima"),
		NoTelp:       c.Query("no_telp"),
		Page:         p,
	}
	rows, err := h.uc.ListAddresses(authUser(c).ID, filter)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", paginated(rows, p))
}

func (h *Handler) getAddress(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	row, err := h.uc.GetAddress(authUser(c).ID, id)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", row)
}

func (h *Handler) createAddress(c *gin.Context) {
	var req domain.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	id, err := h.uc.CreateAddress(authUser(c).ID, req)
	if err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	ok(c, "POST", id)
}

func (h *Handler) updateAddress(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	var req domain.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	if err := h.uc.UpdateAddress(authUser(c).ID, id, req); err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	ok(c, "PUT", "")
}

func (h *Handler) deleteAddress(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "DELETE", err)
		return
	}
	if err := h.uc.DeleteAddress(authUser(c).ID, id); err != nil {
		fail(c, http.StatusBadRequest, "DELETE", err)
		return
	}
	ok(c, "DELETE", "")
}

func (h *Handler) listCategories(c *gin.Context) {
	p := page(c).Normalize()
	filter := repository.CategoryFilter{
		Search:       c.Query("search"),
		NamaCategory: c.Query("nama_category"),
		Page:         p,
	}
	rows, err := h.uc.ListCategories(filter)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", paginated(rows, p))
}

func (h *Handler) getCategory(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	row, err := h.uc.GetCategory(id)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", row)
}

func (h *Handler) createCategory(c *gin.Context) {
	var req domain.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	id, err := h.uc.CreateCategory(authUser(c).IsAdmin, req)
	if err != nil {
		fail(c, http.StatusUnauthorized, "POST", err)
		return
	}
	ok(c, "POST", id)
}

func (h *Handler) updateCategory(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	var req domain.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	if err := h.uc.UpdateCategory(authUser(c).IsAdmin, id, req); err != nil {
		fail(c, http.StatusUnauthorized, "PUT", err)
		return
	}
	ok(c, "PUT", "")
}

func (h *Handler) deleteCategory(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "DELETE", err)
		return
	}
	if err := h.uc.DeleteCategory(authUser(c).IsAdmin, id); err != nil {
		fail(c, http.StatusUnauthorized, "DELETE", err)
		return
	}
	ok(c, "DELETE", "")
}

func (h *Handler) myStore(c *gin.Context) {
	store, err := h.uc.MyStore(authUser(c).ID)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", store)
}

func (h *Handler) getStoreDynamic(c *gin.Context) {
	if c.Param("id_toko") == "my" {
		h.jwt.Middleware()(c)
		if c.IsAborted() {
			return
		}
		h.myStore(c)
		return
	}
	h.getStore(c)
}

func (h *Handler) getStore(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id_toko"))
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	store, err := h.uc.GetStore(id)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", store)
}

func (h *Handler) listStores(c *gin.Context) {
	p := page(c).Normalize()
	rows, err := h.uc.ListStores(c.Query("nama"), p)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", paginated(rows, p))
}

func (h *Handler) updateStore(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id_toko"))
	if err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	if err := h.uc.UpdateStore(authUser(c).ID, id, c.PostForm("nama_toko"), multipartFiles(c, "photo")); err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	ok(c, "PUT", "")
}

func (h *Handler) listProducts(c *gin.Context) {
	p := page(c).Normalize()
	filter := repository.ProductFilter{
		NamaProduk: c.Query("nama_produk"),
		CategoryID: queryUint(c, "category_id"),
		TokoID:     queryUint(c, "toko_id"),
		MinHarga:   queryInt64(c, "min_harga"),
		MaxHarga:   queryInt64(c, "max_harga"),
		Page:       p,
	}
	rows, err := h.uc.ListProducts(filter)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", paginated(rows, p))
}

func (h *Handler) getProduct(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	row, err := h.uc.GetProduct(id)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", row)
}

func (h *Handler) createProduct(c *gin.Context) {
	in, err := productInput(c, true)
	if err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	id, err := h.uc.CreateProduct(authUser(c).ID, in)
	if err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	ok(c, "POST", id)
}

func (h *Handler) updateProduct(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	in, err := productInput(c, false)
	if err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	if err := h.uc.UpdateProduct(authUser(c).ID, id, in); err != nil {
		fail(c, http.StatusBadRequest, "PUT", err)
		return
	}
	ok(c, "PUT", "")
}

func (h *Handler) deleteProduct(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "DELETE", err)
		return
	}
	if err := h.uc.DeleteProduct(authUser(c).ID, id); err != nil {
		fail(c, http.StatusBadRequest, "DELETE", err)
		return
	}
	ok(c, "DELETE", "")
}

func (h *Handler) listTransactions(c *gin.Context) {
	p := page(c).Normalize()
	filter := repository.TransactionFilter{
		MethodBayar: c.Query("method_bayar"),
		MinTotal:    firstQueryInt64(c, "min_total", "min_harga"),
		MaxTotal:    firstQueryInt64(c, "max_total", "max_harga"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		Page:        p,
	}
	rows, err := h.uc.ListTransactions(authUser(c).ID, filter)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", paginated(rows, p))
}

func (h *Handler) getTransaction(c *gin.Context) {
	id, err := usecase.ParseUint(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	row, err := h.uc.GetTransaction(authUser(c).ID, id)
	if err != nil {
		fail(c, http.StatusBadRequest, "GET", err)
		return
	}
	ok(c, "GET", row)
}

func (h *Handler) createTransaction(c *gin.Context) {
	var req domain.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	id, err := h.uc.CreateTransaction(authUser(c).ID, req)
	if err != nil {
		fail(c, http.StatusBadRequest, "POST", err)
		return
	}
	ok(c, "POST", id)
}

func (h *Handler) listProvinces(c *gin.Context) {
	data, err := h.provcity.ListProvinces(c.Query("search"), queryInt(c, "page"), queryInt(c, "limit"))
	if err != nil {
		fail(c, http.StatusBadGateway, "get", err)
		return
	}
	okMessage(c, "Succeed to get data", data)
}

func (h *Handler) listCities(c *gin.Context) {
	data, err := h.provcity.ListCities(c.Param("prov_id"), c.Query("search"), queryInt(c, "page"), queryInt(c, "limit"))
	if err != nil {
		fail(c, http.StatusBadGateway, "get", err)
		return
	}
	okMessage(c, "Succeed to get data", data)
}

func (h *Handler) detailProvince(c *gin.Context) {
	data, err := h.provcity.DetailProvince(c.Param("prov_id"))
	if err != nil {
		fail(c, http.StatusBadGateway, "get", err)
		return
	}
	okMessage(c, "Succeed to get data", data)
}

func (h *Handler) detailCity(c *gin.Context) {
	data, err := h.provcity.DetailCity(c.Param("city_id"))
	if err != nil {
		fail(c, http.StatusBadGateway, "get", err)
		return
	}
	okMessage(c, "Succeed to get data", data)
}

func authResponse(user domain.User, token string) domain.AuthResponse {
	return domain.AuthResponse{
		Nama:         user.Nama,
		NoTelp:       user.NoTelp,
		TanggalLahir: user.TanggalLahir,
		Tentang:      user.Tentang,
		Pekerjaan:    user.Pekerjaan,
		Email:        user.Email,
		IDProvinsi:   user.IDProvinsi,
		IDKota:       user.IDKota,
		Token:        token,
	}
}

func productInput(c *gin.Context, requireAll bool) (usecase.ProductInput, error) {
	in := usecase.ProductInput{
		NamaProduk: c.PostForm("nama_produk"),
		Photos:     multipartFiles(c, "photos"),
	}
	if val, ok := c.GetPostForm("category_id"); ok && val != "" {
		parsed, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return usecase.ProductInput{}, err
		}
		in.CategoryID = uint(parsed)
		in.HasCategoryID = true
	} else if requireAll {
		return usecase.ProductInput{}, fmt.Errorf("category_id is required")
	}
	if val, ok := c.GetPostForm("harga_reseller"); ok && val != "" {
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return usecase.ProductInput{}, err
		}
		in.HargaReseller = parsed
		in.HasHargaReseller = true
	} else if requireAll {
		return usecase.ProductInput{}, fmt.Errorf("harga_reseller is required")
	}
	if val, ok := c.GetPostForm("harga_konsumen"); ok && val != "" {
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return usecase.ProductInput{}, err
		}
		in.HargaKonsumen = parsed
		in.HasHargaKonsumen = true
	} else if requireAll {
		return usecase.ProductInput{}, fmt.Errorf("harga_konsumen is required")
	}
	if val, ok := c.GetPostForm("stok"); ok && val != "" {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return usecase.ProductInput{}, err
		}
		in.Stok = parsed
		in.HasStok = true
	} else if requireAll {
		return usecase.ProductInput{}, fmt.Errorf("stok is required")
	}
	if val, ok := c.GetPostForm("deskripsi"); ok {
		in.Deskripsi = val
		in.HasDeskripsi = true
	}
	return in, nil
}

func page(c *gin.Context) repository.PageFilter {
	return repository.PageFilter{Page: queryInt(c, "page"), Limit: queryInt(c, "limit")}
}

func queryInt(c *gin.Context, key string) int {
	val, _ := strconv.Atoi(c.Query(key))
	return val
}

func queryUint(c *gin.Context, key string) uint {
	val, _ := strconv.ParseUint(c.Query(key), 10, 64)
	return uint(val)
}

func queryInt64(c *gin.Context, key string) int64 {
	val, _ := strconv.ParseInt(c.Query(key), 10, 64)
	return val
}

func firstQueryInt64(c *gin.Context, keys ...string) int64 {
	for _, key := range keys {
		if c.Query(key) == "" {
			continue
		}
		return queryInt64(c, key)
	}
	return 0
}

func paginated(data any, p repository.PageFilter) gin.H {
	p = p.Normalize()
	return gin.H{
		"data":  data,
		"page":  p.Page,
		"limit": p.Limit,
	}
}
