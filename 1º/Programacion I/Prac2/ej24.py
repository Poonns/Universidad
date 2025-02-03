n = int(input('Introduce un numero: '))
divisor = 0
suma = 0
for i in range (1, n+1):
    if n % i== 0:
        divisor += 1

print(f'El numero de divisores de {n} es: {divisor}')