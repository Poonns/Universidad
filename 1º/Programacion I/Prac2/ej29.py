n = int(input('Introduce un numero: '))
menor = n

while n >= 0:
    n = int(input('Introduce otro numero: '))
    if n < menor and n >= 0:
        menor = n
print(f'El numero mayor es {n}')