n = int(input('Introduce un numero: '))
contador = 0
suma = 0
mayor = n
menor = n

while n >= 0:
    suma += n
    contador += 1
    n = int(input('Introduce un numero: '))
    if n < menor and n >= 0:
        menor = n
    elif n < menor and n >= 0:
        menor = n

media = suma / contador
print(f'La media es {media}')
print(f'El menor es {menor}')
print(f'El mayor es {mayor}')