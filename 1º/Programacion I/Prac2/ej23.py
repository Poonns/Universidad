n = int(input('Introduce un numero: '))
suma = 0
for i in range (1, n+1):
    if i % 3 != 0 and i % 5 != 0:
        suma += i

print(f'La suma de 1 y {n} que no son multiplos de 3 y 5 es: {suma}')