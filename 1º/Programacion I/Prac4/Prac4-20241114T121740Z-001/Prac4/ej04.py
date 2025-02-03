
def leer_lista_enteros():
    lista = []
    cadena = input('Introduce un número: ')
    while cadena != ' ':
        numero = int(cadena)
        lista.append(numero)
        cadena = input('Introduce otro número: ')
    return lista

#Creamos una lista para guardar los valores de la función y luego hacemos otra para introducir los valores modificados
lista = leer_lista_enteros()
cuadrados = []
for numero in lista:
    oper = numero ** 2
    cuadrados.append(oper)

print(f'Cuadrados = {cuadrados}')
print(f'Números leidos = {lista}')

