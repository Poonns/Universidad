public class Modelo{
    private Color[] colores = {Color.RED, Color.GREEN, Color. BLACK, Color.BLUE};
    private int i = 0;

    public Color getColorActual(){
        return colores[i];
    }
    
    public void siguienteColor(){
        i++;
        if(i >= colores.length){
            i = 0;
        }
    }

    public void anteriorColor(){
        i--;
        if(i < 0){
            i = colores.length -1;
        }
    }
}