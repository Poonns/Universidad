n = int(input('Introduce un numero: '))
divisor = 0
suma = 0
for i in range (1, n+1):
    if n % i== 0:
        divisor += 1
        suma += i

print(f'La suma de los divisores de {n} es: {suma}')