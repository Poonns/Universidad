class Cuenta:
    def __init__(self, saldo):
        self.saldo = saldo

    def ingresar(self, importe):
        self.saldo += importe

    def reintegrar(self, importe):
        self.saldo -= importe

    def consultar_estado(self):
        return self.saldo
    
cuenta = Cuenta(int(input('Saldo que quieres introducir en la cuenta: ')))

while True:
    print('1. Hacer un ingreso')
    print('2. Pedir un reintegro')
    print('3. Consultar estado')
    print('4. Salir')

    opcion = int(input('Selecciona una opción: '))


    if opcion == 1:
        cuenta.ingresar(float(input('Introduce cuantos euros quieres ingresar: ')))
    elif opcion == 2:
        cuenta.reintegrar(float(input('Introduce cuentos euros quieres retirar: ')))
    elif opcion == 3:
        print(cuenta.consultar_estado())
    elif opcion == 4:
        print('Saliendo del programa...')
        break
    else:
        print('Opcion no valida')

