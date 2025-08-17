package main

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Number:    randRange.Intn(10000000),
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	require := require.New(t)
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(err, "не удалось открыть БД")
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(err, "ошибка при добавлении посылки")
	require.NotZero(id, "ID посылки не был присвоен")

	// Получение посылки
	gotParcel, err := store.Get(id)
	require.NoError(err, "ошибка при получении посылки")
	// Проверка всех полей
	require.Equal(parcel.Number, gotParcel.Number, "не совпадает номер посылки")
	require.Equal(parcel.Client, gotParcel.Client, "не совпадает клиент")
	require.Equal(parcel.Status, gotParcel.Status, "не совпадает статус")
	require.Equal(parcel.Address, gotParcel.Address, "не совпадает адрес")
	require.Equal(parcel.CreatedAt, gotParcel.CreatedAt, "не совпадает дата создания")

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(err, "ошибка при удалении посылки")

	// Проверка удаления
	_, err = store.Get(id)
	require.Error(err, "посылка должна быть удалена")
	require.True(errors.Is(err, sql.ErrNoRows), "ожидается ошибка отсутствия строк")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatal("failed to add parcel", err)
	}
	if id == 0 {
		t.Error("parcel ID is not returned correctly")
	}
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Fatal("failed to set address", err)
	}
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	gotParcel, err := store.Get(id)
	if err != nil {
		t.Fatal("failed to get parcel", err)
	}
	if gotParcel.Address != newAddress {
		t.Errorf("address does not match: expected %v, got %v", newAddress, gotParcel.Address)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel() // настройте подключение к БД

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatal("failed to add parcel", err)
	}
	if id == 0 {
		t.Error("parcel ID is not returned correctly")
	}
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusRegistered
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Fatal("failed to set status", err)
	}
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	gotParcel, err := store.Get(id)
	if err != nil {
		t.Fatal("failed to get parcel", err)
	}
	if gotParcel.Status != ParcelStatusRegistered {
		t.Errorf("status does not match: expected %v, got %v", ParcelStatusRegistered, gotParcel.Status)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	// настройте подключение к БД

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatal("failed to add parcel", err)
		}
		if id == 0 {
			t.Error("parcel ID is not returned correctly")
		} // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatal("failed to get parcels by client", err)
	}
	if len(storedParcels) != len(parcels) {
		t.Errorf("number of parcels does not match: expected %v, got %v", len(parcels), len(storedParcels))
	}

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		parcel, ok := parcelMap[parcel.Number]
		if !ok {
			t.Errorf("parcel not found in parcelMap: %v", parcel)
		}
		if parcel.Client != client {
			t.Errorf("client does not match: expected %v, got %v", client, parcel.Client)
		}
	}
}
