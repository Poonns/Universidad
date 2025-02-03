def lee_lista_enteros():
    lista = []
    cadena = input('Introduce un numero: ')
    while cadena != '':
        numero =int(cadena)
        lista.append(numero)
        cadena = input('Introduce otro numero: ')
    return lista

lista = lee_lista_enteros()
print(f'Lista leida {lista}')
print('Ve contestando con números enteros, o una cadena vacia para acabar...')
resultado_suma = input('¿Qué suma he de buscar en dos posiciones consecutivas?')

while resultado_suma != '':
    resultado_suma = int(resultado_suma)
    for i in range(len(lista)-1):
        suma = lista[i] + lista[i+1]
        if suma == resultado_suma:
            print(f'{lista[i]} + {lista[i+1]} = {resultado_suma}')

    resultado_suma = input('¿Qué suma he de buscar en dos posiciones consecutivas? ')

print('¡Adiós!')
