public class Controlador{
    private Vista vista;
    int xIni = 0;
    int yIni = 0;
    int xFin = 0;
    int yFin = 0;

    public Controlador(Vista v){
        vista = v;

        vista.getPanel().addMouseListener(new MouseAdapter(){
            @Override
            public void mousePressed(MouseEvent e){
                xIni = e.getX();
                yIni = e.getY();
            }
            @Override
            public void mouseReleased(MouseEvent e){
                xFin = e.getX();
                yFin = e.getY();

                vista.setPos(xIni, xFin, yIni, yFin);
            }
        });

        vista.addKeyListener(new KeyAdapter(){
            @Override
            public void keyTyped(KeyEvent e){
                vista.setTecla(e.getKeyChar());
            }
        });
    }
}