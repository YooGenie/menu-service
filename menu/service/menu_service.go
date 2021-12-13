package service

import (
	"context"
	requestDto "menu-service/dto/request"
	"menu-service/dto/response"
	"menu-service/menu/entity"
	"menu-service/menu/mapper"
	"menu-service/menu/repository"
	"sync"
)

var (
	menuServiceOnce     sync.Once
	menuServiceInstance *menuService
)

func MenuService() *menuService {
	menuServiceOnce.Do(func() {
		menuServiceInstance = &menuService{}
	})

	return menuServiceInstance
}

type menuService struct {
}

func (menuService) CreateMenu(ctx context.Context, creation requestDto.MenuCreate) (err error) {
	newMenu, err := mapper.NewMenu(creation)
	if err != nil {
		return
	}
	if err = repository.MenuRepository().Create(ctx, &newMenu); err != nil {
		return err
	}
	return err

}

func (menuService) GetMenuById(ctx context.Context, Id int64) (entity.Menu, error) {
	return repository.MenuRepository().FindById(ctx, Id)
}

func (menuService) GetMenu(ctx context.Context, pageable response.Pageable) ([]entity.Menu, int64, error) {
	return repository.MenuRepository().FindAll(ctx, pageable)
}

func (menuService) UpdateMenu(ctx context.Context, menuMake requestDto.MenuCreate) (int64, error) {
	//menu, err := repository.MenuRepository().FindById(ctx, menuMake.Id)
	//if err != nil {
	//	return 0, err
	//}
	//
	//menu.UpdateMenu(ctx, menuMake)
	//
	//if err := repository.MenuRepository().Update(ctx, &menu); err != nil {
	//	return 0, err
	//}

	//return menu.Id, nil
	return 0, nil
}

func (menuService) DeleteMenu(ctx context.Context, Id int64) error {
	menu, err := repository.MenuRepository().FindById(ctx, Id)
	if err != nil {
		return err
	}

	return repository.MenuRepository().Delete(ctx, &menu)
}
