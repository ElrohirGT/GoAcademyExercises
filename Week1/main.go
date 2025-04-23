package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Estudiante struct {
	nombre       string
	edad         uint
	carrera      string
	calificacion map[string]float32
	promedio     float32
}

// Crear un programa para gestionar estudiantes en una universidad. El sistema debe permitir:
// Registrar estudiantes con su información básica. (Nombre, Edad, Carrera, Calificaciones) Tip: Las calificaciones que sea un map en dónde el key es la materia y el value es la calificación.
// * Almacenar sus calificaciones por materia.
// * Calcular el promedio de calificaciones de cada estudiante.
// * Determinar si aprueban o no usando condicionales (> 70 pasan).
// * Agrupar a los estudiantes por carrera.
// * Permitir modificar sus calificaciones usando punteros.
// * Calcular estadísticas generales del sistema.

func (estudiante *Estudiante) CambiarCalificacion(materia string, calificacion float32) {
	estudiante.calificacion[materia] = calificacion
}

func (estudiante *Estudiante) Imprimir() string {
	b := strings.Builder{}
	b.WriteString("{ ")
	b.WriteString(estudiante.nombre)
	b.WriteString(", ")

	b.WriteString(estudiante.carrera)
	b.WriteString(", ")

	for curso, nota := range estudiante.calificacion {
		b.WriteString(curso)
		b.WriteString(": ")
		b.WriteString(strconv.FormatFloat(float64(nota), 'f', 2, 32))
		b.WriteString(" ")
	}

	b.WriteString("}")

	return b.String()
}

func initData() []Estudiante {
	flavio := Estudiante{nombre: "Flavio", edad: 21, carrera: "Computación", calificacion: map[string]float32{
		"Matematica":   95,
		"Compiladores": 75,
	}}
	jose := Estudiante{nombre: "Jose", edad: 20, carrera: "Arquitectura", calificacion: map[string]float32{
		"Matematica": 95,
		"Diseño":     90,
	}}
	maria := Estudiante{nombre: "Maria", edad: 23, carrera: "Mecatrónica", calificacion: map[string]float32{
		"Matematica": 95,
		"Diseño 3D":  85,
	}}
	luis := Estudiante{nombre: "Luis", edad: 22, carrera: "Mecatrónica", calificacion: map[string]float32{
		"Matematica": 70,
		"Diseño 3D":  65,
	}}

	return []Estudiante{flavio, jose, maria, luis}
}

func computeAverages(seccion *[]Estudiante) {
	for i, estudiante := range *seccion {

		var promedio float32 = 0.0
		var cuenta float32 = 0.0
		for _, nota := range estudiante.calificacion {
			promedio += nota
			cuenta += 1
		}
		promedio /= cuenta

		(*seccion)[i].promedio = promedio
		// estudiante.Promedio = promedio
	}
}

func getApprovedStudents(seccion *[]Estudiante) []Estudiante {
	approved := []Estudiante{}
	for _, estudiante := range *seccion {
		var promedio float32 = 0.0
		var cuenta float32 = 0.0
		for _, nota := range estudiante.calificacion {
			promedio += nota
			cuenta += 1
		}
		promedio /= cuenta

		if promedio >= 70 {
			approved = append(approved, estudiante)
		}
	}

	return approved
}

func groupByCareer(seccion *[]Estudiante) map[string][]Estudiante {
	estudiantesPorCarrera := map[string][]Estudiante{}
	for _, estudiante := range *seccion {
		if _, existe := estudiantesPorCarrera[estudiante.carrera]; !existe {
			estudiantesPorCarrera[estudiante.carrera] = []Estudiante{}
		}

		estudiantesPorCarrera[estudiante.carrera] = append(estudiantesPorCarrera[estudiante.carrera], estudiante)
	}

	return estudiantesPorCarrera
}

func getGlobalAverage(seccion *[]Estudiante) float32 {
	var promedioGlobal float32 = 0.0
	var cuentaGlobal float32 = 0.0
	for _, estudiante := range *seccion {

		var promedio float32 = 0.0
		var cuenta float32 = 0.0
		for _, nota := range estudiante.calificacion {
			promedio += nota
			cuenta += 1
		}
		promedio /= cuenta

		promedioGlobal += promedio
		cuentaGlobal += 1
	}
	promedioGlobal /= cuentaGlobal

	return promedioGlobal
}

func main() {
	seccion := initData()

	fmt.Printf("Promedio de calificaciones de cada estudiante:\n")
	computeAverages(&seccion)
	for _, estudiante := range seccion {
		fmt.Printf("Estudiante: %s, Promedio: %f\n", estudiante.nombre, estudiante.promedio)
	}

	fmt.Printf("Estudiantes que aprueban:\n")
	approved := getApprovedStudents(&seccion)
	for _, estudiante := range approved {
		fmt.Printf("%s\n", estudiante.nombre)
	}

	fmt.Printf("Agrupando a los estudiantes por carrera:\n")
	estudiantesPorCarrera := groupByCareer(&seccion)
	for carrera, estudiantes := range estudiantesPorCarrera {
		fmt.Printf("%s:", carrera)
		for _, estudiante := range estudiantes {
			fmt.Printf("%s ", estudiante.Imprimir())
		}
		fmt.Println()
	}

	fmt.Printf("Cambiando calificaciones de Luis:\n")
	fmt.Printf("ANTES: %s\n", seccion[3].Imprimir())
	seccion[3].CambiarCalificacion("Diseño 3D", 85)
	seccion[3].CambiarCalificacion("Matematica", 95)
	fmt.Printf("DESPUES: %s\n", seccion[3].Imprimir())

	fmt.Printf("Estadísticas generales del sistema:\n")
	promedioGlobal := getGlobalAverage(&seccion)
	fmt.Printf("Promedio de la seccion: %f\n", promedioGlobal)
}
