name_file = input('Introduce el nombre del archivo: ')

contador_digitos = 0
contador_letras = 0

file = open (name_file, 'r')
contenido = file.read()

for caracter in contenido:
    if '0' <= caracter <= '9':
        contador_digitos += 1
    elif 'A' <= caracter <= 'Z' or 'a' < caracter <= 'z':
        contador_letras += 1

print(contador_digitos)
print(contador_letras)

file.close()