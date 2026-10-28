# Compilador y opciones
CC = g++
CCFLAGS = -Wall 

# Archivos para test_tabla
OBJ_TEST = test_tabla.o alumno.o
TARGET_TEST = test_tabla

# Archivos para costes
OBJ_COSTES = costes.o alumno.o
TARGET_COSTES = costes

# Regla principal
all: $(TARGET_TEST) $(TARGET_COSTES)

# Regla para test_tabla
$(TARGET_TEST): $(OBJ_TEST)
	$(CC) $(CCFLAGS) -o $(TARGET_TEST) $(OBJ_TEST)

# Regla para costes
$(TARGET_COSTES): $(OBJ_COSTES)
	$(CC) $(CCFLAGS) -o $(TARGET_COSTES) $(OBJ_COSTES)

# Reglas para generar los .o
test_tabla.o: test_tabla.cpp tabla.h alumno.h
	$(CC) $(CCFLAGS) -c test_tabla.cpp

costes.o: costes.cpp tabla.h alumno.h abb.h
	$(CC) $(CCFLAGS) -c costes.cpp

alumno.o: alumno.cpp alumno.h
	$(CC) $(CCFLAGS) -c alumno.cpp

# Limpiar archivos intermedios (.o) y el ejecutable generado
clean:
	rm -f $(OBJ_TEST) $(OBJ_COSTES) $(TARGET_TEST) $(TARGET_COSTES)