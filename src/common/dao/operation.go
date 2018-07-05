package dao

import (
	"git/inspursoft/board/src/common/model"

	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func AddOperation(operation model.Operation) (int64, error) {
	o := orm.NewOrm()

	operation.CreationTime = time.Now()

	operationID, err := o.Insert(&operation)
	if err != nil {
		if err == orm.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return operationID, err
}

func GetOperation(operation model.Operation, fieldNames ...string) (*model.Operation, error) {
	o := orm.NewOrm()
	err := o.Read(&operation, fieldNames...)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &operation, err
}

func UpdateOperation(operation model.Operation, fieldNames ...string) (int64, error) {
	o := orm.NewOrm()
	operation.UpdateTime = time.Now()
	fieldNames = append(fieldNames, "update_time")
	operationID, err := o.Update(&operation, fieldNames...)
	if err != nil {
		if err == orm.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return operationID, err
}

func GetPaginatedOperations(query model.OperationParam, pageIndex int, pageSize int, orderField string, orderAsc int) (*model.PaginatedOperations, error) {
	sql, params := generateOperationSQL(query)
	var err error

	pagination := &model.Pagination{
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	pagination.TotalCount, err = getTotalRecordCount(sql, params)
	if err != nil {
		return nil, err
	}
	sql += getOrderSQL(orderField, orderAsc) + ` limit ?, ?`
	params = append(params, pagination.GetPageOffset(), pagination.PageSize)

	logs.Debug("GetPaginatedOperations: +%+v", pagination.String())

	operationList, err := queryOperations(sql, params)
	if err != nil {
		return nil, err
	}

	return &model.PaginatedOperations{
		OperationList: operationList,
		Pagination:    pagination,
	}, nil
}

func queryOperations(sql string, params []interface{}) ([]*model.Operation, error) {
	o := orm.NewOrm()
	operationList := make([]*model.Operation, 0)
	_, err := o.Raw(sql, params).QueryRows(&operationList)

	if err != nil {
		return nil, err
	}
	return operationList, nil
}

func generateOperationSQL(query model.OperationParam) (string, []interface{}) {
	sql := `select * from operation o where 1 = 1 `

	params := make([]interface{}, 0)
	params = append(params)

	if query.Object != "" {
		params = append(params, "%"+query.Object+"%")
		sql += ` and o.object_type like ? `
	}
	if query.Action != "" {
		params = append(params, "%"+query.Action+"%")
		sql += ` and o.action like ? `
	}
	if query.User != "" {
		params = append(params, "%"+query.User+"%")
		sql += ` and o.user_name like ? `
	}
	if query.Status != "" {
		params = append(params, "%"+query.Status+"%")
		sql += ` and o.status like ? `
	}
	if query.Fromdate != "" {
		params = append(params, query.Fromdate)
		sql += ` and o.creation_time >= ? `
	}
	if query.Todate != "" {
		params = append(params, query.Todate)
		sql += ` and o.creation_time <= ? `
	}

	return sql, params
}
