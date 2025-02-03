
num = int(input('Introduce un numero: '))
contador = 0
divisores = 2

while contador < num:
        suma_divisores = 0
        
        for i in range (2, divisores):
            if divisores % i == 0:
                suma_divisores += i

        if suma_divisores > divisores:
            print(divisores)
            contador += 1
        
        divisores += 1

