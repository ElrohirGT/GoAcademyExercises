package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"goacademy.com/week2/payments"
)

// Se requiere la construcción de el núcleo de un sistema de pagos para PayFlex, una pasarela que admite múltiples métodos de pagocomo tarjetas de crédito, transferencias bancarias, criptomonedas, etc.
// Tu tarea es desarrollar una arquitectura modular y robusta que permita registrar e integrar nuevos métodos de pago fácilmente, manejar tiempos de espera en la validación de transacciones, y proteger el sistema de errores críticos durante la ejecución.
//
// 1. Definir una interfaz PaymentMethod para representar cualquier método de pago.
// 2. Crear una interfaz extendida Refundable para métodos que soportan devoluciones.
// 3. Implementar un plugin CardPayment que simule un pago con tarjeta, debe de tener un Delay que simule un pago y manejar un rango para que que sea valida el pago. Ej: Los pagos solo son validos desde 1USD a 1000USD. Para la política de devolución sólo basta quemar un mensaje que indique que el reembolso será realizado en X tiempo.
// 4. Crear un orquestador genérico PaymentProcessor[T] que ejecute un método de pago y detecte si el método también implementa Refundable, si implementa Refundable basta con que se imprima la política de reembolso.
// 5. Usar context.Context para abortar pagos que tarden demasiado.
// 6. Usar defer y recover() para manejar errores críticos.
// 7. Organizar el proyecto por paquetes.
// 8. Separar claramente la lógica de negocio (plugins), de la lógica de orquestación y de la presentación.

func main() {
	log.SetOutput(io.Discard)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in main with value:", r)
		}
	}()

	// panic("Hola!")

	fmt.Println("Creando pagos...")
	payment := payments.NewCardPayment(50)
	payment2 := payments.NewCardPayment(500)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	processor := payments.PaymentProcessor{}
	fmt.Println("Ejecutando el pago 1...")
	if processor.Process(ctx, payment) {
		fmt.Println("Pago 1 ejecutado con éxito!")
	} else {
		fmt.Println("Pago 1 FALLO!")
	}

	fmt.Println("Ejecutando el pago 2...")
	if processor.Process(ctx, payment2) {
		fmt.Println("Pago 2 ejecutado con éxito!")
	} else {
		fmt.Println("Pago 2 FALLO!")
	}
}
