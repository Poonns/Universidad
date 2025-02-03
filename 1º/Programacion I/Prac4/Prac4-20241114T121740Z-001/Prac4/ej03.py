lista = []
pares = []
impares = []
cadena = int(input('Introduce un número: '))

while cadena > 0:
    lista.append(cadena)
    cadena = int(input('Introduce otro numero: '))

for numero in lista:
    if numero % 2 == 0:
        pares.append(numero)
    else: 
        impares.append(numero)

print(f'Pares = {pares}')
print(f'Impares = {impares}')