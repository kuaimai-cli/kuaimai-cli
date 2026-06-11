# 员工管理（/staff）

## 接口

| operation | 说明 |
|-----------|------|
| `staff-query` | 员工列表（分页） |
| `staff-show-edit-staff-shop` | 编辑员工店铺权限 |
| `staff-gyl-show-create-staff` | 新增员工权限列表 |
| `staff-gyl-show-edit-staff` | 编辑员工权限列表 |
| `staff-save-or-update-staff-info` | 保存/更新员工（写） |

## staff-query

```bash
kuaimai-cli web call scm.staff-query \
  --body '{"queryStaffName":"张三","pageNo":1,"pageSize":20}' \
  --output json --no-color
```

常用字段：`queryStaffName`（姓名/账号/手机）、`pageNo`、`pageSize`、`useSupplyChain`（0/1）。

## staff-show-edit-staff-shop

```bash
kuaimai-cli web call scm.staff-show-edit-staff-shop \
  --body '{"staffId":136283493321216}' \
  --output json --no-color
```

## 写操作

`staff-save-or-update-staff-info` 为写接口，须 `--dry-run --verbose` 预览后再提交。
