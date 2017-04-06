package Controllers

import (
	"github.com/chapzin/parse-efd-fiscal/Models"
	"github.com/chapzin/parse-efd-fiscal/Models/Bloco0"
	"github.com/chapzin/parse-efd-fiscal/Models/BlocoC"
	"github.com/chapzin/parse-efd-fiscal/Models/BlocoH"
	"github.com/chapzin/parse-efd-fiscal/Models/NotaFiscal"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/tealeg/xlsx"
	"strconv"
	"sync"
	"time"
)

func ProcessarFatorConversao(db gorm.DB, wg *sync.WaitGroup) {
	time.Sleep(1 * time.Second)
	color.Green("Começo Processa Fator de Conversao %s", time.Now())
	db.Exec("DELETE FROM reg_0220 WHERE fat_conv=1")
	var fator []Bloco0.Reg0220
	db.Where("feito = ?", 0).Find(&fator)
	for _, vFator := range fator {
		c170 := []BlocoC.RegC170{}
		db.Where("cod_item = ? and unid = ? and dt_ini = ? and dt_fin = ?", vFator.CodItem, vFator.UnidConv, vFator.DtIni, vFator.DtFin).Find(&c170)
		for _, vC170 := range c170 {
			nvC170 := BlocoC.RegC170{}
			nvC170.Qtd = vC170.Qtd * vFator.FatConv
			nvC170.Unid = vFator.UnidCod
			db.Table("reg_c170").Where("id = ? and cod_item = ?", vC170.ID, vC170.CodItem).Update(&nvC170)
			nvFator := Bloco0.Reg0220{}
			nvFator.Feito = "1"
			db.Table("reg_0220").Where("id = ?", vFator.ID).Update(&nvFator)
		}
	}
	color.Green("Fim Processa Fator de Conversao %s", time.Now())
	wg.Done()
}

func DeletarItensNotasCanceladas(db gorm.DB, dtIni string, dtFin string, wg *sync.WaitGroup) {
	color.Green("Começo Deleta itens notas canceladas %s", time.Now())
	var c100 []BlocoC.RegC100
	db.Where("cod_sit <> ? and dt_ini >= ? and dt_ini <= ? ", "00", dtIni, dtFin).Find(&c100)
	for _, v := range c100 {
		//fmt.Println(v.NumDoc)
		var nota []NotaFiscal.NotaFiscal
		db.Where("ch_n_fe = ?", v.ChvNfe).Find(&nota)
		for _, v2 := range nota {
			db.Where("nota_fiscal_id =?", v2.ID).Delete(NotaFiscal.Item{})
		}
	}
	db.Exec("DELETE FROM items WHERE deleted_at is not null")
	color.Green("Fim deleta itens notas canceladas %s", time.Now())
	wg.Done()
}

func PopularReg0200(db gorm.DB, wg *sync.WaitGroup) {
	time.Sleep(1 * time.Second)
	color.Green("Comeco popula reg0200 %s", time.Now())
	var reg0200 []Bloco0.Reg0200
	db.Where("tipo_item=00").Select("distinct cod_item,descr_item,tipo_item,unid_inv").Find(&reg0200)
	for _, v := range reg0200 {
		inv2 := Models.Inventario{
			Codigo:    v.CodItem,
			Descricao: v.DescrItem,
			Tipo:      v.TipoItem,
			UnidInv:   v.UnidInv,
			Ncm:       v.CodNcm,
		}
		db.NewRecord(inv2)
		db.Create(&inv2)

	}
	wg.Done()
	color.Green("Fim popula reg0200 %s", time.Now())
}

func PopularItensXmls(db gorm.DB, wg *sync.WaitGroup) {
	color.Green("Comeco popula Itens Xmls %s", time.Now())
	var items []NotaFiscal.Item
	db.Select("distinct codigo,descricao").Find(&items)
	for _, v := range items {
		var inventario Models.Inventario
		db.Where("codigo=?", v.Codigo).First(&inventario)
		if inventario.Codigo == "" {
			inv2 := Models.Inventario{
				Codigo:    v.Codigo,
				Descricao: v.Descricao,
			}
			db.NewRecord(inv2)
			db.Create(&inv2)
		}
	}
	wg.Done()
	color.Green("Fim popula xmls %s", time.Now())

}

func PopularInventarios(AnoInicial int, AnoFinal int, wg *sync.WaitGroup, db gorm.DB) {
	time.Sleep(1 * time.Second)
	color.Green("Começo popula Inventario %s", time.Now())
	qtdAnos := AnoFinal - AnoInicial
	ano1 := AnoInicial
	qtdAnos = qtdAnos + 1
	ano := 0
	for qtdAnos >= 0 {
		qtdAnos = qtdAnos - 1
		ano = ano + 1
		var regH010 []BlocoH.RegH010
		var inv []Models.Inventario
		AnoInicialString := strconv.Itoa(ano1)
		db.Where("dt_ini= ?", AnoInicialString+"-02-01").Find(&regH010)
		db.Find(&inv)
		for _, vInv := range inv {
			for _, vH010 := range regH010 {
				if vH010.CodItem == vInv.Codigo {
					inv3 := Models.Inventario{}
					switch ano {
					case 1:
						inv3.InvFinalAno1 = vH010.Qtd
						inv3.VlInvAno1 = vH010.VlUnit
					case 2:
						inv3.InvFinalAno2 = vH010.Qtd
						inv3.VlInvAno2 = vH010.VlUnit
					case 3:
						inv3.InvFinalAno3 = vH010.Qtd
						inv3.VlInvAno3 = vH010.VlUnit
					case 4:
						inv3.InvFinalAno4 = vH010.Qtd
						inv3.VlInvAno4 = vH010.VlUnit
					case 5:
						inv3.InvFinalAno5 = vH010.Qtd
						inv3.VlInvAno5 = vH010.VlUnit
					case 6:
						inv3.InvFinalAno6 = vH010.Qtd
						inv3.VlInvAno6 = vH010.VlUnit

					}
					db.Table("inventarios").Where("codigo = ?", vH010.CodItem).Update(&inv3)
				}
			}
		}
		ano1 = ano1 + 1
	}
	color.Green("Fim popula inventario %s", time.Now())
	wg.Done()
}

func PopularEntradas(AnoInicial int, AnoFinal int, wg *sync.WaitGroup, db gorm.DB) {
	time.Sleep(1 * time.Second)
	color.Green("Começo popula entradas %s", time.Now())
	qtdAnos := AnoFinal - AnoInicial
	ano1 := AnoInicial
	ano := 0
	for qtdAnos >= 0 {
		qtdAnos = qtdAnos - 1
		ano = ano + 1
		AnoInicialString := strconv.Itoa(ano1)

		dtIni := AnoInicialString + "-01-01"
		dtFin := AnoInicialString + "-12-31"

		var inv []Models.Inventario
		var c170 []BlocoC.RegC170
		var itens []NotaFiscal.Item

		db.Find(&inv)
		db.Where("entrada_saida = ? and dt_ini >= ? and dt_fin <= ? ", "0", dtIni, dtFin).Find(&c170)
		db.Where("cfop < 3999 and dt_emit >= ? and dt_emit <= ?", dtIni, dtFin).Find(&itens)
		for _, vInv := range inv {
			var qtd_tot = 0.0
			var vl_tot = 0.0
			for _, vc170 := range c170 {
				if vc170.CodItem == vInv.Codigo {
					qtd_tot = qtd_tot + vc170.Qtd
					vl_tot = vl_tot + vc170.VlItem
				}
			}

			for _, vitens := range itens {
				if vitens.Codigo == vInv.Codigo {
					qtd_tot = qtd_tot + vitens.Qtd
					vl_tot = vl_tot + vitens.VTotal
				}
			}
			inv2 := Models.Inventario{}
			switch ano {
			case 1:
				inv2.EntradasAno2 = qtd_tot
				inv2.VlTotalEntradasAno2 = vl_tot
			case 2:
				inv2.EntradasAno3 = qtd_tot
				inv2.VlTotalEntradasAno3 = vl_tot
			case 3:
				inv2.EntradasAno4 = qtd_tot
				inv2.VlTotalEntradasAno4 = vl_tot
			case 4:
				inv2.EntradasAno5 = qtd_tot
				inv2.VlTotalEntradasAno5 = vl_tot
			case 5:
				inv2.EntradasAno6 = qtd_tot
				inv2.VlTotalEntradasAno6 = vl_tot

			}
			db.Table("inventarios").Where("codigo = ?", vInv.Codigo).Update(&inv2)
		}

		ano1 = ano1 + 1
	}
	color.Green("Fim popula entradas %s", time.Now())
	wg.Done()
}

func PopularSaidas(AnoInicial int, AnoFinal int, wg *sync.WaitGroup, db gorm.DB) {
	time.Sleep(2 * time.Second)
	color.Green("Comeco popula saidas %s", time.Now())
	qtdAnos := AnoFinal - AnoInicial
	ano1 := AnoInicial
	ano := 0
	for qtdAnos >= 0 {
		qtdAnos = qtdAnos - 1
		ano = ano + 1
		AnoInicialString := strconv.Itoa(ano1)

		dtIni := AnoInicialString + "-01-01"
		dtFin := AnoInicialString + "-12-31"

		var inv []Models.Inventario
		var itens []NotaFiscal.Item
		var c425 []BlocoC.RegC425

		db.Find(&inv)
		db.Where("cfop > 3999 and cfop <> 5929 and cfop <> 6929 and dt_emit >= ? and dt_emit <= ?", dtIni, dtFin).Find(&itens)
		db.Where("dt_ini >= ? and dt_ini <= ?", dtIni, dtFin).Find(&c425)

		for _, vInv := range inv {
			var qtd_saida = 0.0
			var vl_tot_saida = 0.0
			for _, vItens := range itens {
				if vItens.Codigo == vInv.Codigo {
					qtd_saida = qtd_saida + vItens.Qtd
					vl_tot_saida = vl_tot_saida + vItens.VTotal
				}
			}
			for _, vc425 := range c425 {
				if vc425.CodItem == vInv.Codigo {
					qtd_saida = qtd_saida + vc425.Qtd
					vl_tot_saida = vl_tot_saida + vc425.VlItem
				}
			}
			inv3 := Models.Inventario{}
			switch ano {
			case 1:
				inv3.SaidasAno2 = qtd_saida
				inv3.VlTotalSaidasAno2 = vl_tot_saida
			case 2:
				inv3.SaidasAno3 = qtd_saida
				inv3.VlTotalSaidasAno3 = vl_tot_saida
			case 3:
				inv3.SaidasAno4 = qtd_saida
				inv3.VlTotalSaidasAno4 = vl_tot_saida
			case 4:
				inv3.SaidasAno5 = qtd_saida
				inv3.VlTotalSaidasAno5 = vl_tot_saida
			case 5:
				inv3.SaidasAno6 = qtd_saida
				inv3.VlTotalSaidasAno6 = vl_tot_saida

			}
			db.Table("inventarios").Where("codigo = ?", vInv.Codigo).Update(&inv3)

		}
		ano1 = ano1 + 1
	}
	color.Green("Fim popula saidas %s", time.Now())
	wg.Done()
}

/*
 * fazer uma refactory completo recursive
func ProcessarDiferencas(db gorm.DB) {
	db.Exec("Delete from inventarios where inv_inicial=0 and entradas=0 and vl_total_entradas=0 and saidas=0 and vl_total_saidas=0 and inv_final=0")
	var inv []Models.Inventario
	var reg0200 []Bloco0.Reg0200
	db.Select("distinct cod_item,descr_item,tipo_item,unid_inv").Find(&reg0200)
	db.Find(&inv)
	for _, vInv := range inv {
		inv3 := Models.Inventario{}
		// Calculando as diferencas
		diferencas := (vInv.InvInicial + vInv.Entradas) - (vInv.Saidas + vInv.InvFinal)

		// Calculando o valor unitário de entrada
		if vInv.VlTotalEntradas > 0 && vInv.Entradas > 0 {
			inv3.VlUnitEnt = vInv.VlTotalEntradas / vInv.Entradas
		} else if vInv.VlTotalEntradas == 0 && vInv.Entradas == 0 && vInv.VlInvIni > 0 {
			inv3.VlUnitEnt = vInv.VlInvIni
		} else if vInv.VlTotalEntradas == 0 && vInv.Entradas == 0 && vInv.VlInvIni == 0 && vInv.VlInvFin > 0 {
			inv3.VlUnitEnt = vInv.VlInvFin
		} else {
			inv3.VlUnitEnt = 1
		}

		// Calculando o valor unitário de saida
		if vInv.VlTotalSaidas > 0 && vInv.Saidas > 0 {
			inv3.VlUnitSai = vInv.VlTotalSaidas / vInv.Saidas
		} else if vInv.VlTotalSaidas == 0 && vInv.Saidas == 0 && vInv.VlInvIni > 0 {
			inv3.VlUnitSai = vInv.VlInvIni
		} else {
			inv3.VlUnitSai = 0
		}

		// Criando Sugestao de novo inventário
		if diferencas >= 0 {
			// Novo inventario final somando diferencas
			nvInvFin := diferencas + vInv.InvFinal
			inv3.SugInvFinal = nvInvFin
			inv3.SugVlInvFinal = nvInvFin * inv3.VlUnitEnt
		} else {
			inv3.SugInvFinal = vInv.InvFinal
			inv3.SugVlInvFinal = inv3.SugInvFinal * inv3.VlUnitEnt
		}
		if diferencas < 0 {
			// Caso negativo adiciona ao inventario inicial
			nvInvIni := (diferencas * -1) + vInv.InvInicial
			inv3.SugInvInicial = nvInvIni
			inv3.SugVlInvInicial = nvInvIni * inv3.VlUnitEnt
		} else {
			// Caso nao seja negativo mantenha o inventario anterior
			inv3.SugInvInicial = vInv.InvInicial
			inv3.SugVlInvInicial = inv3.SugInvInicial * inv3.VlUnitEnt
		}

		// Zera o produto quando inventario inicial e final forem iguais
		if inv3.SugInvInicial == inv3.SugInvFinal {
			inv3.SugInvFinal = 0
			inv3.SugInvInicial = 0
			inv3.SugVlInvFinal = 0
			inv3.SugVlInvInicial = 0
		}
		// Adicionando Tipo e unidade de medida no inventario
		for _, v0200 := range reg0200 {
			if v0200.CodItem == vInv.Codigo {
				inv3.Tipo = v0200.TipoItem
				inv3.UnidInv = v0200.UnidInv
			}
		}
		inv3.Diferencas = diferencas
		db.Table("inventarios").Where("codigo = ?", vInv.Codigo).Update(&inv3)
	}
	// Deleta tudo tipo de inventario que nao seja material de revenda
	db.Exec("Delete from inventarios where tipo <> '00'")
}

*/
func ExcelAdd(db gorm.DB, sheet *xlsx.Sheet) {
	var inv []Models.Inventario
	db.Find(&inv)
	for _, vInv := range inv {
		ExcelItens(sheet, vInv)
	}
}

func ColunaAdd(linha *xlsx.Row, string string) {
	cell := linha.AddCell()
	cell.Value = string
}

func ColunaAddFloat(linha *xlsx.Row, valor float64) {
	cell := linha.AddCell()
	cell.SetFloat(valor)
}
func ColunaAddFloatDif(linha *xlsx.Row, valor float64) {
	cell := linha.AddCell()

	var style = xlsx.NewStyle()
	if valor < 0 {
		style.Fill = *xlsx.NewFill("solid", "00FA8072", "00FA8072")
	} else if valor > 0 {
		style.Fill = *xlsx.NewFill("solid", "0087CEFA", "0087CEFA")
	} else {
		style.Fill = *xlsx.NewFill("solid", "009ACD32", "009ACD32")
	}
	cell.SetStyle(style)

	cell.SetFloat(valor)
}

func ExcelItens(sheet *xlsx.Sheet, inv Models.Inventario) {
	menu := sheet.AddRow()
	// Produto
	ColunaAdd(menu, inv.Codigo)
	ColunaAdd(menu, inv.Descricao)
	ColunaAdd(menu, inv.Tipo)
	ColunaAdd(menu, inv.UnidInv)
	ColunaAdd(menu, inv.Ncm)
	// Ano 1
	ColunaAddFloat(menu, inv.InvFinalAno1)
	ColunaAddFloat(menu, inv.VlInvAno1)
	// Ano 2
	ColunaAddFloat(menu, inv.EntradasAno2)
	ColunaAddFloat(menu, inv.VlTotalEntradasAno2)
	ColunaAddFloat(menu, inv.VlUnitEntAno2)
	ColunaAddFloat(menu, inv.SaidasAno2)
	ColunaAddFloat(menu, inv.VlTotalSaidasAno2)
	ColunaAddFloat(menu, inv.VlUnitSaiAno2)
	ColunaAddFloat(menu, inv.MargemAno2)
	ColunaAddFloat(menu, inv.InvFinalAno2)
	ColunaAddFloat(menu, inv.VlInvAno2)
	ColunaAddFloatDif(menu, inv.DiferencasAno2)
	// Ano 3
	ColunaAddFloat(menu, inv.EntradasAno3)
	ColunaAddFloat(menu, inv.VlTotalEntradasAno3)
	ColunaAddFloat(menu, inv.VlUnitEntAno3)
	ColunaAddFloat(menu, inv.SaidasAno3)
	ColunaAddFloat(menu, inv.VlTotalSaidasAno3)
	ColunaAddFloat(menu, inv.VlUnitSaiAno3)
	ColunaAddFloat(menu, inv.MargemAno3)
	ColunaAddFloat(menu, inv.InvFinalAno3)
	ColunaAddFloat(menu, inv.VlInvAno3)
	ColunaAddFloatDif(menu, inv.DiferencasAno3)
	// Ano 4
	ColunaAddFloat(menu, inv.EntradasAno4)
	ColunaAddFloat(menu, inv.VlTotalEntradasAno4)
	ColunaAddFloat(menu, inv.VlUnitEntAno4)
	ColunaAddFloat(menu, inv.SaidasAno4)
	ColunaAddFloat(menu, inv.VlTotalSaidasAno4)
	ColunaAddFloat(menu, inv.VlUnitSaiAno4)
	ColunaAddFloat(menu, inv.MargemAno4)
	ColunaAddFloat(menu, inv.InvFinalAno4)
	ColunaAddFloat(menu, inv.VlInvAno4)
	ColunaAddFloatDif(menu, inv.DiferencasAno4)
	// Ano 5
	ColunaAddFloat(menu, inv.EntradasAno5)
	ColunaAddFloat(menu, inv.VlTotalEntradasAno5)
	ColunaAddFloat(menu, inv.VlUnitEntAno5)
	ColunaAddFloat(menu, inv.SaidasAno5)
	ColunaAddFloat(menu, inv.VlTotalSaidasAno5)
	ColunaAddFloat(menu, inv.VlUnitSaiAno5)
	ColunaAddFloat(menu, inv.MargemAno5)
	ColunaAddFloat(menu, inv.InvFinalAno5)
	ColunaAddFloat(menu, inv.VlInvAno5)
	ColunaAddFloatDif(menu, inv.DiferencasAno5)
	// Ano 6
	ColunaAddFloat(menu, inv.EntradasAno6)
	ColunaAddFloat(menu, inv.VlTotalEntradasAno6)
	ColunaAddFloat(menu, inv.VlUnitEntAno6)
	ColunaAddFloat(menu, inv.SaidasAno6)
	ColunaAddFloat(menu, inv.VlTotalSaidasAno6)
	ColunaAddFloat(menu, inv.VlUnitSaiAno6)
	ColunaAddFloat(menu, inv.MargemAno6)
	ColunaAddFloat(menu, inv.InvFinalAno6)
	ColunaAddFloat(menu, inv.VlInvAno6)
	ColunaAddFloatDif(menu, inv.DiferencasAno6)
}

func ExcelMenu(sheet *xlsx.Sheet) {
	menu := sheet.AddRow()
	// Produtos
	ColunaAdd(menu, "Codigo")
	ColunaAdd(menu, "Descricao")
	ColunaAdd(menu, "Tipo")
	ColunaAdd(menu, "Unid_inv")
	ColunaAdd(menu, "NCM")
	// Ano 1
	ColunaAdd(menu, "InvFinalAno1")
	ColunaAdd(menu, "VlInvAno1")
	// Ano 2
	ColunaAdd(menu, "EntradasAno2")
	ColunaAdd(menu, "VlTotalEntradasAno2")
	ColunaAdd(menu, "VlUnitEntAno2")
	ColunaAdd(menu, "SaidasAno2")
	ColunaAdd(menu, "VlTotalSaidasAno2")
	ColunaAdd(menu, "VlUnitSaidaAno2")
	ColunaAdd(menu, "MargemAno2")
	ColunaAdd(menu, "InvFinalAno2")
	ColunaAdd(menu, "VlInvAno2")
	ColunaAdd(menu, "DiferencasAno2")
	// Ano 3
	ColunaAdd(menu, "EntradasAno3")
	ColunaAdd(menu, "VlTotalEntradasAno3")
	ColunaAdd(menu, "VlUnitEntAno3")
	ColunaAdd(menu, "SaidasAno3")
	ColunaAdd(menu, "VlTotalSaidasAno3")
	ColunaAdd(menu, "VlUnitSaidaAno3")
	ColunaAdd(menu, "MargemAno3")
	ColunaAdd(menu, "InvFinalAno3")
	ColunaAdd(menu, "VlInvAno3")
	ColunaAdd(menu, "DiferencasAno3")
	// Ano 4
	ColunaAdd(menu, "EntradasAno4")
	ColunaAdd(menu, "VlTotalEntradasAno4")
	ColunaAdd(menu, "VlUnitEntAno4")
	ColunaAdd(menu, "SaidasAno4")
	ColunaAdd(menu, "VlTotalSaidasAno4")
	ColunaAdd(menu, "VlUnitSaidaAno4")
	ColunaAdd(menu, "MargemAno4")
	ColunaAdd(menu, "InvFinalAno4")
	ColunaAdd(menu, "VlInvAno4")
	ColunaAdd(menu, "DiferencasAno4")
	// Ano 5
	ColunaAdd(menu, "EntradasAno5")
	ColunaAdd(menu, "VlTotalEntradasAno5")
	ColunaAdd(menu, "VlUnitEntAno5")
	ColunaAdd(menu, "SaidasAno5")
	ColunaAdd(menu, "VlTotalSaidasAno5")
	ColunaAdd(menu, "VlUnitSaidaAno5")
	ColunaAdd(menu, "MargemAno5")
	ColunaAdd(menu, "InvFinalAno5")
	ColunaAdd(menu, "VlInvAno5")
	ColunaAdd(menu, "DiferencasAno5")
	// Ano 6
	ColunaAdd(menu, "EntradasAno6")
	ColunaAdd(menu, "VlTotalEntradasAno6")
	ColunaAdd(menu, "VlUnitEntAno6")
	ColunaAdd(menu, "SaidasAno6")
	ColunaAdd(menu, "VlTotalSaidasAno6")
	ColunaAdd(menu, "VlUnitSaidaAno6")
	ColunaAdd(menu, "MargemAno6")
	ColunaAdd(menu, "InvFinalAno6")
	ColunaAdd(menu, "VlInvAno6")
	ColunaAdd(menu, "DiferencasAno6")
}

/*
func CriarH010InvInicial(ano int, db gorm.DB) {
	ano = ano - 1
	anoString := strconv.Itoa(ano)
	var inv []Models.Inventario
	db.Find(&inv)
	f, err := os.Create("SpedInvInicial.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, vInv := range inv {
		if vInv.SugInvInicial > 0 {
			r0200 := Bloco0.Reg0200{
				Reg:       "0200",
				CodItem:   vInv.Codigo,
				DescrItem: vInv.Descricao,
				UnidInv:   vInv.UnidInv,
				TipoItem:  vInv.Tipo,
			}
			aliqicms := tools.FloatToStringSped(r0200.AliqIcms)
			linha := "|" + r0200.Reg + "|" + r0200.CodItem + "|" + r0200.DescrItem + "|" +
				r0200.CodBarra + "|" + r0200.CodAntItem + "|" + r0200.UnidInv + "|" + r0200.TipoItem +
				"|" + r0200.CodNcm + "|" + r0200.ExIpi + "|" + r0200.CodGen + "|" + r0200.CodLst +
				"|" + aliqicms + "|\r\n"
			f.WriteString(linha)
			f.Sync()
		}
	}
	linha := "|H005|3112" + anoString + "|1726778,31|01|\r\n"
	f.WriteString(linha)
	f.Sync()

	for _, vInv2 := range inv {
		if vInv2.SugInvInicial > 0 {
			sugVlUnit := vInv2.SugVlInvInicial / vInv2.SugInvInicial
			h010 := BlocoH.RegH010{
				Reg:     "H010",
				CodItem: vInv2.Codigo,
				Unid:    vInv2.UnidInv,
				Qtd:     vInv2.SugInvInicial,
				VlUnit:  sugVlUnit,
				VlItem:  vInv2.SugVlInvInicial,
				IndProp: "0",
			}
			linha := "|" + h010.Reg + "|" + h010.CodItem + "|" + h010.Unid + "|" +
				tools.FloatToStringSped(h010.Qtd) + "|" + tools.FloatToStringSped(h010.VlUnit) +
				"|" + tools.FloatToStringSped(h010.VlItem) + "|" + h010.IndProp + "|" + h010.CodPart +
				"|" + h010.CodCta + "|" + tools.FloatToStringSped(h010.VlItemIr) + "|\r\n"
			f.WriteString(linha)
			f.Sync()
		}

	}

	w := bufio.NewWriter(f)
	w.Flush()
}

func CriarH010InvFinal(ano int, db gorm.DB) {
	anoString := strconv.Itoa(ano)
	var inv []Models.Inventario
	db.Find(&inv)
	f, err := os.Create("SpedInvFinal.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, vInv := range inv {
		if vInv.SugInvFinal > 0 {
			r0200 := Bloco0.Reg0200{
				Reg:       "0200",
				CodItem:   vInv.Codigo,
				DescrItem: vInv.Descricao,
				UnidInv:   vInv.UnidInv,
				TipoItem:  vInv.Tipo,
			}
			aliqicms := tools.FloatToStringSped(r0200.AliqIcms)
			linha := "|" + r0200.Reg + "|" + r0200.CodItem + "|" + r0200.DescrItem + "|" +
				r0200.CodBarra + "|" + r0200.CodAntItem + "|" + r0200.UnidInv + "|" + r0200.TipoItem +
				"|" + r0200.CodNcm + "|" + r0200.ExIpi + "|" + r0200.CodGen + "|" + r0200.CodLst +
				"|" + aliqicms + "|\r\n"
			f.WriteString(linha)
			f.Sync()
		}
	}
	linha := "|H005|3112" + anoString + "|1726778,31|01|\r\n"
	f.WriteString(linha)
	f.Sync()

	for _, vInv2 := range inv {
		if vInv2.SugInvFinal > 0 {
			sugVlUnit := vInv2.SugVlInvFinal / vInv2.SugInvFinal
			h010 := BlocoH.RegH010{
				Reg:     "H010",
				CodItem: vInv2.Codigo,
				Unid:    vInv2.UnidInv,
				Qtd:     vInv2.SugInvFinal,
				VlUnit:  sugVlUnit,
				VlItem:  vInv2.SugVlInvFinal,
				IndProp: "0",
			}
			linha := "|" + h010.Reg + "|" + h010.CodItem + "|" + h010.Unid + "|" +
				tools.FloatToStringSped(h010.Qtd) + "|" + tools.FloatToStringSped(h010.VlUnit) +
				"|" + tools.FloatToStringSped(h010.VlItem) + "|" + h010.IndProp + "|" + h010.CodPart +
				"|" + h010.CodCta + "|" + tools.FloatToStringSped(h010.VlItemIr) + "|\r\n"
			f.WriteString(linha)
			f.Sync()
		}

	}

	w := bufio.NewWriter(f)
	w.Flush()
}

*/
