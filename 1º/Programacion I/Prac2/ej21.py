n = int(input('Introduce un numero: '))
for i in range (1, n+1):
    if i ** 2 < n:
        print(f'El cuadrado de {i} es {i ** 2}')